# Frontend — интеграция с MAX

## Как мини-приложение подключается к MAX

1. `index.html` грузит скрипт бриджа:
   `<script src="https://st.max.ru/js/max-web-app.js"></script>`.
2. Скрипт создаёт глобальный объект `window.WebApp` — он же MAX Bridge.
3. Вне MAX объект отсутствует → `getWebApp()` возвращает `null`,
   приложение должно деградировать честно (см. App.tsx).

## Типы (src/lib/max.ts)

`MaxWebApp` — основное поле `initData: string` (подписанная строка для
бекенда), плюс `initDataUnsafe` (распарсенные данные, **только для чтения
в UI**, не доверять для логики), `platform`, `colorScheme`, `themeParams`,
методы `ready/expand/close/sendData/onEvent/offEvent`, `HapticFeedback`.

`MaxUser` — `id`, `first_name`, `last_name`, `username`, `photo_url`.

## Хелперы

| Функция         | Что делает                                        |
|-----------------|---------------------------------------------------|
| `getWebApp()`   | `WebApp` или `null` вне MAX                       |
| `isInsideMax()` | Открыто ли внутри MAX                             |
| `getInitData()` | Строка initData для `POST /auth` (вне MAX — `""`) |

## Жизненный цикл мини-приложения

1. При монтировании вызвать `webApp.ready()` — снимает загрузочный экран MAX.
2. `webApp.expand()` — раскрыть на весь экран.
3. Авторизация: `api.auth(getInitData())` → бекенд upsert'ит пользователя.
4. Выход из приложения — `webApp.close()`.

## Тема

`colorScheme` (`light`/`dark`) и `themeParams` приходят из клиента MAX.
Сейчас приложение стилизовано под тёмную тему константой; при добавлении
светлой — подписаться через `webApp.onEvent('themeChanged', ...)` и
перекладывать схему в CSS-переменные.

## Безопасность

- `initDataUnsafe` — это данные клиента, им нельзя доверять: любую логику
  и отображение чужих данных валидировать на бекенде.
- Бекенд пока не проверяет подпись initData (см. `docs/backend/api.md`) —
  значит и фронт не должен рассчитывать, что `user.id` «настоящий».

## Стилизация под нативный UI

Опционально можно подключить библиотеку компонентов
[`@maxhub/max-ui`](https://github.com/max-messenger/max-ui)
(`npm i @maxhub/max-ui`), провайдер `MaxUI` + компоненты `Panel`, `Button` и
т.д. В базовой структуре не используется.
