.PHONY: run build test vet tidy swag

run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

swag:
	swag init -g cmd/server/main.go --parseDependency --parseDepth 2 -o docs

modernize:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix ./...
