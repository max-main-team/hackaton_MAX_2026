# T-23 — Реферальная система (бек)

- **Статус:** ⬜
- **Спека:** docs/common/features.md (фича 3), docs/max/mini-app-setup.md (диплинки)
- **Зависимости:** T-03

## Что сделать

- Миграция: `ALTER TABLE users ADD COLUMN referrer_id BIGINT REFERENCES users (id)`,
  `ADD COLUMN referrer_at TIMESTAMPTZ`.
- В `POST /auth`: если в initData есть `start_param` вида `ref_<id>` и
  `referrer_id` ещё не установлен и `id != ref_id` → записать реферера.
- `GET /my/referrals` → `{count, items: [{id, first_name, joined_at}]}`.
- `GET /me`/`/auth` ответ дополняется `referral_code` (=`ref_<id>`) — фронт
  собирает диплинк `https://max.ru/t599_hakaton_max_bot?startapp=ref_<id>`.

## Критерии приёмки

- [ ] Юзер 1 зашёл по `startapp=ref_42` → у него `referrer_id=42`.
- [ ] Повторные заходы не меняют реферера; самоссылка игнорируется.
- [ ] `/my/referrals` у юзера 42 показывает юзера 1.
- [ ] go build/vet/gofmt — чисто.
