# Валидация initData (подпись стартовых параметров)

Источник: https://dev.max.ru/docs/webapps/validation

Это закрывает TODO из T-01: проверка подписи initData на бекенде.

## Откуда берётся initData

- Клиент MAX при каждом запуске передаёт мини-аппу параметры в URL-фрагменте:
  `https://app.example.com#WebAppData=<закодированная строка>&WebAppPlatform=web&WebAppVersion=...`
- В JS то же самое доступно как `window.WebApp.initData` — строка вида
  `chat=...&ip=...&user=...&query_id=...&auth_date=...&hash=...`
  (пары `key=value` через `&`, значения URL-кодированы).
- `auth_date` — Unix timestamp **в секундах**. Рекомендация MAX: считать
  данные невалидными спустя **1 час** после `auth_date`.

## Структура параметров

| Ключ | Значение |
|------|----------|
| `query_id` | ID сессии |
| `ip` | IP пользователя (опционально) |
| `auth_date` | Unix-секунды выдачи данных |
| `hash` | Подпись параметров |
| `user` | URL-encoded JSON: `{id, first_name, last_name, username, language_code, photo_url}` |
| `chat` | URL-encoded JSON: `{id, type: 'DIALOG'\|'CHAT'\|'CHANNEL'}` |
| `start_param` | payload из диплинка `?startapp=` |

## Алгоритм проверки (10 шагов)

Вход: `initData` (строка из `window.WebApp.initData`), `BOT_TOKEN`.

1. Разбить `initData` по `&` на пары `key=value` (каждый параметр
   встречается ровно один раз).
2. Убедиться, что `hash` встречается ровно один раз; сохранить
   оригинальный хеш и **исключить** его из массива.
3. URL-декодировать все значения (`url.ParseQuery` в Go делает это сам).
4. Отсортировать пары по ключам алфавитно `a → z`.
5. Склеить в строку `key1=value1\nkey2=value2` (разделитель `\n`, 0x0A) —
   это `launch_params`.
6. `secret_key = HMAC_SHA256(key="WebAppData", message=BOT_TOKEN)`
   — строка `"WebAppData"` является ключом HMAC, токен бота — сообщением.
7. `hash = hex(HMAC_SHA256(key=secret_key, message=launch_params))`.
8. Сравнить hex-строку с оригинальным `hash` (constant-time compare).
   Совпало → данные подлинные.

⚠️ **Секрет — это токен бота**, к которому привязано мини-приложение
(не отдельный ключ). В проекте это env `MAX_BOT_TOKEN`.

## Эталонная реализация на Go

```go
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func VerifyInitData(rawInitData, botToken string) error {
	values, err := url.ParseQuery(rawInitData)
	if err != nil {
		return fmt.Errorf("parse initData: %w", err)
	}
	origHash := values.Get("hash")
	if origHash == "" {
		return fmt.Errorf("hash is missing")
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
		return fmt.Errorf("hash mismatch")
	}
	return nil
}
```

Заметки по реализации:

- `url.ParseQuery` сам URL-декодирует значения (шаг 3) — но собирать
  `launch_params` нужно из **декодированных** значений, что `values.Get(k)`
  и возвращает.
- Сравнение хешей — только через `hmac.Equal` (защита от timing-атак).
- Свежесть: `auth_date` не старше 1 часа (рекомендация MAX; в dev можно
  ослабить, договорившись внутри команды).
- В dev-режиме (пустой `MAX_BOT_TOKEN`) проверка пропускается с warning —
  это режим хакатона, в проде токен обязателен.

## Дополнительно: проверка телефона из `requestContact()`

`hash == HMAC_SHA256(authDate + phone + userId, botToken)`,
пары через `\n`, `phone` без `+`. Позволяет убедиться, что пользователь
поделился номером, привязанным к MAX.
