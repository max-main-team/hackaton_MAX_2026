# Bot API — необходимый минимум

Источники: https://dev.max.ru/docs-api , `/docs-api/methods/POST/messages`,
`/docs-api/use-cases/sending-messages/keyboard`

Нужен для T-12 (воркер еженедельного опроса) и вообще любых сообщений
от бота.

## Базовые вещи

- **Домен:** `https://platform-api2.max.ru` (старый `platform-api.max.ru`
  не использовать).
- **Авторизация:** заголовок `Authorization: <access_token>` (токен бота).
  Через query-параметры токен передавать больше нельзя.
- Токен берётся на платформе business.max.ru → Чат-боты → ⋮ → Настройки
  (или у бота «MAX для бизнеса», команда «Получить токен»).
- На сервере должен быть добавлен в доверенные **сертификат Минцифры**.
- У нас токен бота = `MAX_BOT_TOKEN` (он же секрет валидации initData,
  см. `init-data.md`).

## Отправка сообщения: POST /messages

```bash
curl -X POST "https://platform-api2.max.ru/messages?user_id={user_id}" \
  -H "Authorization: {MAX_BOT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"text": "Привет!", "notify": true}'
```

- `user_id` — в диалог пользователю (наш случай: `users.id` из initData!),
  либо `chat_id` — в групповой чат/канал.
- Тело: `text` (до 4000 символов), `attachments[]`, `format`
  (`markdown`/`html`), `notify` (push; для каналов только `true`),
  `disable_link_preview`.
- Результат: объект `message` (в нём `mid` — нужен для `shareMaxContent`).
- Коды: 200 ок, 401 токен, 500 сервер.

**Лимит: не более 2 сообщений/сек в один диалог** — при рассылке ставить
очередь/задержку.

Редактирование: `PUT /messages`; удаление: `DELETE /messages`.

## Inline-клавиатура (кнопки в сообщении)

Вложение `type: "inline_keyboard"`:

```json
{
  "text": "Ваш подбор ещё актуален?",
  "attachments": [
    {
      "type": "inline_keyboard",
      "payload": {
        "buttons": [
          [
            { "type": "open_app", "text": "Ответить в приложении" }
          ]
        ]
      }
    }
  ]
}
```

Типы кнопок:

| Тип | Действие | Заметки |
|-----|----------|---------|
| `callback` | сервер шлёт событие `message_callback` (по webhook/long polling) | отвечать через `POST /answers` |
| `link` | открыть URL (≤2048 симв.) | можно диплинк мини-аппа |
| `open_app` | **открыть мини-приложение бота** | идеально для опроса (T-12) |
| `message` | отправить боту заданный текст | |
| `request_contact` | поделиться телефоном (с `hash` для проверки) | |
| `request_geo_location` | запросить геолокацию | пригодится карте вакансий |
| `clipboard` | скопировать `payload` в буфер | |

Геометрия: до 210 кнопок, 30 рядов, до 7 кнопок в ряду (до 3 для
`link`/`open_app`/`request_geo_location`/`request_contact`).

## Получение событий от пользователя

Два способа подписаться на события (нажатия callback, новые сообщения):

- **Webhook:** `POST /subscriptions` (URL куда слать), `GET /subscriptions`,
  `DELETE /subscriptions`.
- **Long Polling:** `GET /updates`.

Для T-12 MVP достаточно **не слушать события вообще**: сообщение с кнопкой
`open_app` → пользователь отвечает внутри мини-аппа →
`POST /my/resume/confirm-activity`. Колбэки понадобятся, если захотим
принимать ответ прямо кнопками в чате.

Ответ на callback-кнопку: `POST /answers`.

## Как понять, от какого бота токен: GET /me

```bash
curl -X GET "https://platform-api2.max.ru/me" \
  -H "Authorization: <MAX_BOT_TOKEN>"
```

Ответ `200` — `User` + `BotInfo`: `user_id`, `first_name` (имя бота),
`username` (никнейм), `is_bot: true`, `description`, `avatar_url`,
`commands`. `401` — токен неверный или отозван.

Практика:
- токен валидации initData обязан принадлежать **тому же боту**, к которому
  привязан мини-апп, — иначе подпись не сойдётся;
- полезно на старте сервера (если `MAX_BOT_TOKEN` задан) дернуть `/me`
  и залогировать имя/ID бота — сразу видно, тем ли ключом деплоились.

## Полезные методы (на будущее)

| Метод | Зачем |
|-------|-------|
| `GET /me` | инфо о боте |
| `PATCH /me/commands` | команды бота |
| `GET /chats`, `GET /chats/{id}` | чаты бота |
| `GET /messages?message_id=` | получить сообщение (например с контактом) |
| `POST /uploads` | загрузка медиа |

## Официальная Go-библиотека ботов

Есть готовая: https://dev.max.ru/docs/chatbots/bots-coding/go —
можно использовать в воркере вместо сырого HTTP, либо обычный
`net/http` с двумя запросами.
