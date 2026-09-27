// Package auth: парсинг и проверка подписи initData (MAX Bridge) и выпуск JWT.
// Алгоритм валидации: https://dev.max.ru/docs/webapps/validation,
// выжимка — docs/max/init-data.md.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrInitDataExpired = errors.New("init data expired")

type InitDataUser struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	PhotoURL     string `json:"photo_url"`
	LanguageCode string `json:"language_code"`
}

func ParseInitData(raw string) (InitDataUser, time.Time, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return InitDataUser{}, time.Time{}, fmt.Errorf("parse init data: %w", err)
	}

	var user InitDataUser
	rawUser := values.Get("user")
	if rawUser == "" {
		return InitDataUser{}, time.Time{}, errors.New("user is missing in init data")
	}
	if err := json.Unmarshal([]byte(rawUser), &user); err != nil {
		return InitDataUser{}, time.Time{}, fmt.Errorf("parse user from init data: %w", err)
	}
	if user.ID == 0 {
		return InitDataUser{}, time.Time{}, errors.New("user id is missing in init data")
	}

	sec, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		return InitDataUser{}, time.Time{}, fmt.Errorf("parse auth_date: %w", err)
	}

	return user, time.Unix(sec, 0), nil
}

func StartParam(raw string) string {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return ""
	}
	return values.Get("start_param")
}

// Verify проверяет подпись initData: подписантом выступает токен бота,
// чьё мини-приложение запущено (см. docs/max/init-data.md).
func Verify(rawInitData, botToken string) error {
	values, err := url.ParseQuery(rawInitData)
	if err != nil {
		return fmt.Errorf("parse init data: %w", err)
	}

	origHash := values.Get("hash")
	if origHash == "" {
		return errors.New("hash is missing")
	}
	if len(values) < 2 {
		return errors.New("not enough params")
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+values.Get(k))
	}
	launchParams := strings.Join(pairs, "\n")

	mac := hmac.New(sha256.New, []byte("WebAppData"))
	mac.Write([]byte(botToken))
	secretKey := mac.Sum(nil)

	mac = hmac.New(sha256.New, secretKey)
	mac.Write([]byte(launchParams))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(origHash)) {
		return errors.New("hash mismatch")
	}
	return nil
}
