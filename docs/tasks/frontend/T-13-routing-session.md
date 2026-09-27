# T-13 — Роутинг, сессия, api-клиенты

- **Статус:** ✅
- **Спека:** `docs/frontend/spec-ui.md` (разделы «Новые зависимости», «Структура», «Роуты»)
- **Зависимости:** T-03 (бек, формат `/auth`) — **один коммит с ним**
- **Разблокирует:** T-14

## Что сделать

- `npm i react-router-dom`.
- `src/lib/session.ts`: getToken/setToken/clearToken (localStorage
  `max_token`), `authHeaders()`, обработка 401 → clearToken → редирект `/`.
- `src/api/`: `client.ts` (fetch + Bearer), `auth.ts` — `auth()` возвращает
  `{ token, user }` (формат T-03, уже реализован), `me()`, `setRole()`;
  `resume.ts`, `company.ts` — типы snake_case по Swagger
  (`https://eclipse-sim.ru/api/docs`).
- `src/main.tsx` → `RouterProvider`; удалить `App.tsx`; заглушки роутов
  по таблице роутов из спеки.

## Критерии приёмки

- [ ] `npm run build && npm run lint` — чисто.
- [ ] Логин через `/auth` сохраняет токен; запросы ходят с Bearer.
- [ ] 401 от бека выбрасывает на `/`.
