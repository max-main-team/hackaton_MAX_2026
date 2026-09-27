# T-21 — Верификация компании через MAX-бота (бек)

- **Статус:** ✅ (реализовано, прод-проверка в рамках финального прогона)
- **Спека:** docs/common/features.md (фича 2), docs/max/bot-api.md (GET /me)
- **Зависимости:** T-07

## Что сделать

- Миграция 000005 (вместе с T-23/T-25 или отдельная): `companies.bot_user_id BIGINT`, `companies.bot_username TEXT`.
- `POST /companies/{id}/verify` — тело `{"bot_token": "..."}`; только участник компании.
  Бек вызывает `GET https://platform-api2.max.ru/me` с этим токеном;
  при `is_bot=true` сохраняет `bot_user_id`/`bot_username` и ставит `verified=true`.
  Невалидный токен / не бот → 400. Секрет Минцифры в контейнере уже настроен.
- `GET /my/companies` и `GET /companies/{id}` возвращают `bot_username`.

## Критерии приёмки

- [ ] Фейковый токен → 400; токен реального бота → verified=true + username сохранён.
- [ ] Бейдж виден в API-ответах.
- [ ] go build/vet/gofmt — чисто.
