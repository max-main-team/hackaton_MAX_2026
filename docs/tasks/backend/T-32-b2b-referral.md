# T-32 — B2B-рефералка «Компания → компания» (бек)

- **Статус:** ✅
- **Спека:** механика из ревью — приглашающая компания получает +N
  приглашений, приглашённая — промо-статус; кандидаты пассивны и в
  рефералке не участвуют
- **Зависимости:** T-03, T-07 (companies)

## Что сделано

- Миграция `000007_b2b_referral`: `companies.referrer_company_id`,
  `referrer_at`, `promo_until`, `invite_quota` (default 3),
  `invite_used`, `users.pending_ref_company_id`.
- Диплинк: `startapp=refc_<company_id>` → в `POST /auth` start_param
  `refc_X` сохраняется пользователю как pending (если компания X
  существует, юзер в ней не состоит и pending ещё пуст).
- При создании компании (`POST /companies` → `ApplyB2BAttribution`,
  одна транзакция): только для **первой** компании создателя —
  `referrer_company_id`, `promo_until = +30 дней`, `invite_quota = 6`;
  приглашавшей компании `invite_quota += 5`. Иначе pending очищается.
- `POST /vacancies/{id}/candidates/{uid}/action {invite}`: списывает
  приглашение (`invite_used < invite_quota`, иначе `409 invite quota
  exceeded`); `skip` не расходует квоту.
- `GET /my/company-referrals` → `{invite_quota, invite_used,
  promo_until, invited: [{id, name, created_at}]}`.

## Критерии приёмки

- [x] Юзер зашёл по `refc_5` → pending=5; создал компанию → у неё
      referrer=5, quota=6, промо; у компании 5 quota 3→8.
- [x] Повторная атрибуция невозможна (referrer_company_id IS NULL +
      первая компания создателя).
- [x] Инвайт списывает квоту; при исчерпании — 409.
- [x] `go build/vet/gofmt`, `go test` — чисто.
