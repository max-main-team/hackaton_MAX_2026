package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"max-miniapp/backend/migrations"
)

type migrationFile struct {
	version   int64
	name      string
	direction string
	path      string
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	if len(entries) == 0 {
		return nil
	}

	files := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		mf, err := parseMigrationName(entry)
		if err != nil {
			return err
		}
		files = append(files, mf)
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].version != files[j].version {
			return files[i].version < files[j].version
		}
		return files[i].direction == "up"
	})

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    BIGINT PRIMARY KEY,
			name       TEXT        NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, mf := range files {
		if mf.direction != "up" {
			continue
		}

		var applied bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`,
			mf.version,
		).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", mf.path, err)
		}
		if applied {
			continue
		}

		sqlBytes, err := migrations.FS.ReadFile(mf.path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", mf.path, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", mf.path, err)
		}
		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", mf.path, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
			mf.version, mf.name,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("mark migration %s as applied: %w", mf.path, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", mf.path, err)
		}
	}

	return nil
}

func parseMigrationName(filename string) (migrationFile, error) {
	trimmed := strings.TrimSuffix(filename, ".sql")
	parts := strings.Split(trimmed, ".")
	if len(parts) != 2 {
		return migrationFile{}, fmt.Errorf("invalid migration name (ожидается 000001_name.up.sql): %s", filename)
	}

	versionStr, name, found := strings.Cut(parts[0], "_")
	if !found {
		return migrationFile{}, fmt.Errorf("invalid migration name (нет версии): %s", filename)
	}
	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		return migrationFile{}, fmt.Errorf("invalid migration version in %s: %w", filename, err)
	}

	direction := parts[1]
	if direction != "up" && direction != "down" {
		return migrationFile{}, fmt.Errorf("invalid migration direction in %s: %s", filename, direction)
	}

	return migrationFile{version: version, name: name, direction: direction, path: filename}, nil
}
