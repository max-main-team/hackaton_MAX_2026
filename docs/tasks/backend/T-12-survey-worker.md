# T-12 — Воркер еженедельного опроса

- **Статус:** ⬜
- **Спека:** `docs/backend/spec-matching.md` (раздел 3)
- **Зависимости:** T-06
- **Разблокирует:** демонстрацию фичи 1.1

## Что сделать

- `internal/worker/survey.go`: горутина с `time.Ticker` (1 час), запуск из `main.go`.
- Выборка: `resumes.is_active = TRUE AND last_confirmed_at < now() - 7 days`.
- Отправка через MAX Bot API (env `MAX_BOT_TOKEN`): сообщение
  «Подбор ещё актуален?» + кнопка `open_app` / deeplink в мини-апп
  (`https://max.ru/t599_hakaton_max_bot?startapp=confirm`).
  Метод: `POST https://platform-api2.max.ru/messages?user_id=<id>`.
- ⚠️ Сертификат Минцифры: у `platform-api2.max.ru` TLS-сертификат
  подписан УЦ Минцифры — системы не доверяют ему по умолчанию. Нужно:
  1) на сервере установить русский корневой CA (`russian_trusted_root_ca_pem.crt`
  с e-trust.gosuslugi.ru → `/usr/local/share/ca-certificates/` →
  `update-ca-certificates`);
  2) добавить CA в backend-образ (`backend/Dockerfile`: COPY +
  `update-ca-certificates`), иначе Go-клиент получит TLS-ошибку.
- Проверка токена: `GET /me` — наш бот «Хакатон МАХ 599»
  (@t599_hakaton_max_bot, user_id 428005502).
- `MAX_BOT_TOKEN` пуст → воркер не стартует, warning.
- `POST /my/resume/confirm-activity` — `{"active": true|false}`:
  true → `last_confirmed_at=now(), is_active=true`; false → `is_active=false`.
- Ошибка отправки одному пользователю не прерывает рассылку.

## Критерии приёмки

- [ ] Пустой токен → в логе warning, рассылки нет.
- [ ] После `confirm-activity` пользователь выпадает из выборки (SQL-проверка).
- [ ] `go build ./... && go vet ./... && gofmt -l .` — чисто.
