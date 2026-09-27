# T-18 — Приглашения кандидата и матч-экран

- **Статус:** ⬜
- **Спека:** `docs/frontend/spec-ui.md` (раздел «Приглашения»)
- **Зависимости:** T-17 (бек: T-11)
- **Разблокирует:** T-20 (демо-прогон)

## Что сделать

- `routes/candidate/invitations.tsx`: `GET /my/invitations`;
  `components/Countdown.tsx` — «осталось N ч» (тик 1 мин) для pending,
  красное «Просрочено» для overdue (кнопки задизейблены),
  responded — показ ответа.
- Кнопки «Принять»/«Отклонить» → `POST /invitations/{id}/respond`.
- Матч-экран после accept: компания (бейдж verified), вакансия,
  контакт рекрутера (`recruiter_contact`), хаптика
  `notificationOccurred('success')`.

## Критерии приёмки

- [ ] pending → accept → матч-экран с контактом; повторный ответ невозможен.
- [ ] overdue показывает «Просрочено», кнопки неактивны.
- [ ] `npm run build && npm run lint` — чисто.
