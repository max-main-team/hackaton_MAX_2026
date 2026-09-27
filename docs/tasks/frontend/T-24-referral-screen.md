# T-24 — Реферальная программа (фронт)

- **Статус:** ✅
- **Спека:** docs/tasks/backend/T-23-referral-api.md, docs/max/mini-app-setup.md
- **Зависимости:** T-14 (бек: T-23)

## Что сделать

- Экран «Пригласить друзей»: `GET /my/referrals` → счётчик приглашённых,
  персональная ссылка `https://max.ru/t599_hakaton_max_bot?startapp=<referral_code>`.
- Кнопки: «Скопировать ссылку» (clipboard/`clipboard` через Bridge),
  «Поделиться» → `WebApp.shareMaxContent({text})` с ссылкой.
- Обработка `start_param`: при заходе по `startapp=ref_<id>` — ничего
  дополнительно не делать на фронте (бек сам парсит initData).

## Критерии приёмки

- [ ] Ссылка копируется и содержит referral_code текущего юзера.
- [ ] Шеринг открывает нативный экран MAX.
- [ ] `npm run build && npm run lint` — чисто.
