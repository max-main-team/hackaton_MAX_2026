# Frontend — архитектура

## Стек

- React 19 + TypeScript (~6.0), строгий tsconfig из шаблона Vite
- Vite 8 (dev-сервер и сборка), плагин `@vitejs/plugin-react`
- oxlint — линтер (`npm run lint`)
- Роутинга и стейт-менеджера нет — добавлять по мере роста
  (рекомендация: react-router + обычные хуки/Zustand)

## Структура

```
frontend/
├── index.html            подключает скрипт MAX Bridge (st.max.ru/js/max-web-app.js)
├── vite.config.ts        dev-прокси /api → http://localhost:8080
└── src/
    ├── main.tsx          точка входа (StrictMode)
    ├── App.tsx           пример: авторизация через initData, показ пользователя
    ├── lib/
    │   ├── max.ts        типы MaxWebApp/MaxUser + getWebApp/isInsideMax/getInitData
    │   └── api.ts        fetch-клиент: API_BASE, ApiError, api.auth/api.health
    └── (styles в App.css / index.css)
```

## Запуск

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173, /api проксируется на :8080
npm run build      # tsc -b && vite build → dist/
npm run lint       # oxlint
```

Бекенд должен быть запущен (`backend/make run`), иначе API-вызовы упадут.

## Переменные окружения

| Переменная         | Назначение                                        |
|--------------------|---------------------------------------------------|
| `VITE_API_BASE_URL`| База API, по умолчанию `/api/v1` (через прокси)   |

## Конвенции

- Именование полей в API-типах — snake_case, как отдаёт бекенд
  (`first_name`, `photo_url`), без маппинга на camelCase.
- Вся работа с MAX — только через хелперы `src/lib/max.ts`, напрямую
  `window.WebApp` в компонентах не трогать.
- Ошибки API — класс `ApiError` со статусом; в UI показываем текст.
- Комментарии в коде не пишем — документация живёт в `docs/`.
