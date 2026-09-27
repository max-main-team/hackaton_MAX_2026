# T-17 — Свайп-лента и список кандидатов

- **Статус:** ✅
- **Спека:** `docs/frontend/spec-ui.md` (разделы «Свайп-лента», «Список кандидатов»)
- **Зависимости:** T-16 (бек: T-09, T-10)
- **Разблокирует:** T-18

## Что сделать

- `components/UserCard.tsx`: имя, фото, title, навыки тегами, score-бейдж.
- `routes/recruiter/feed.tsx` (фича 1.5): `GET /vacancies/{id}/candidates?mode=feed`;
  свайп влево / ✕ → skip, вправо / ♥ → invite (`POST .../action`);
  следующая карточка; `HapticFeedback.impactOccurred('light')` на свайп;
  пусто → «Кандидаты закончились».
- `routes/recruiter/candidate-list.tsx` (фича 1.4): `GET /vacancies/{id}/candidates?mode=list`
  с пагинацией; таблица (имя, title, город, score), сортировка по score,
  клик → разворот карточки с invite/skip.

## Критерии приёмки

- [ ] Invite из ленты и из списка доходит до бека, карточка больше не показывается.
- [ ] Ошибки API отображаются, состояние «пусто» отрисовано.
- [ ] `npm run build && npm run lint` — чисто.
