# MAX Bridge — справочник `window.WebApp`

Источник: https://dev.max.ru/docs/webapps/bridge

Подключение (уже сделано в `frontend/index.html`):

```html
<script src="https://st.max.ru/js/max-web-app.js"></script>
```

Объект `window.WebApp` создаётся клиентом MAX при каждом запуске,
инициализация не нужна. Методы вне MAX отсутствуют — всегда проверять
`getWebApp() !== null`.

## Данные инициализации

| Поле/метод | Тип | Описание |
|------------|-----|----------|
| `initData` | `string` | URL-encoded стартовые параметры для валидации на сервере (см. `init-data.md`) |
| `initDataUnsafe` | `object` | То же в виде JSON. **Для валидации непригоден** — только чтение в UI |
| `platform` | `string` | `ios` \| `android` \| `desktop` \| `web` |
| `version` | `string` | Версия MAX, напр. `25.9.16`; не участвует в хеше |
| `deviceName` | `string` | Человекочитаемое устройство |
| `getLaunchContext()` | `Promise<{entryPoint: 'tabbar'\|'default'}>` | Откуда запущен апп (Android ≥26.19.2, iOS ≥26.20.0) |

Структура `initDataUnsafe`:

```
query_id, ip?, auth_date (unix-сек), hash,
user { id, first_name, last_name, username, language_code, photo_url },
chat { id, type: 'DIALOG'|'CHAT'|'CHANNEL' },
start_param
```

## Экран

| Метод | Описание |
|-------|----------|
| `requestScreenMaxBrightness()` | Макс. яркость на 30 сек |
| `restoreScreenBrightness()` | Вернуть яркость |
| `ScreenCapture.enableScreenCapture()` / `disableScreenCapture()` | Скриншоты |
| `getViewportSize()` | `Promise<{height, width}>` |

## Телефон пользователя

`requestContact()` → `Promise<{phone, authDate, hash}>` (нативный диалог).

Проверка, что номер совпадает с аккаунтом MAX:
`hash == HMAC_SHA256(authDate + phone + userId, botToken)`,
пары `key=value` через `\n`, **phone без `+`** (`7999...`).
Ошибки: `client.request_phone.user_refused_provide_phone_number` (отказ),
`client.request_phone.request_error`.

Для нашей авторизации не нужен — хватит initData. Может пригодиться,
если рекрутеру нужен верифицированный телефон.

## Закрытие

- `enableClosingConfirmation()` / `disableClosingConfirmation()` —
  предупреждение о потере данных при закрытии (включать на формах резюме!).

## Ссылки и файлы

| Метод | Описание |
|-------|----------|
| `openLink(url)` | Внешний браузер (только после клика пользователя) |
| `openMaxLink(url)` | Диплинк `https://max.ru/...` внутри MAX |
| `downloadFile(url, file_name)` | Скачивание; https обязателен, только внутри MAX, не через `href` |

## Шеринг

- `shareContent({text?, link?})` — нативный шеринг ОС (не web).
- `shareMaxContent({text?, link?} \| {mid, chatType})` — шеринг внутри MAX;
  для медиа бот сначала отправляет контент пользователю (`POST /messages`),
  апп получает `mid` и вызывает с `{mid, chatType: 'DIALOG'|'CHAT'}`.

## QR-коды

`openCodeReader(fileSelect = true)` — камера или файл; возвращает распознанную строку.

## Кнопка «Назад»

`BackButton.show() / hide() / isVisible / onClick(cb) / offClick(cb)` —
сохранять ссылку на callback для отписки.

## Хранилища

| Объект | Описание |
|--------|----------|
| `DeviceStorage.setItem/getItem/removeItem/clear` | Локальное хранилище, привязано к пользователю MAX (не web) |
| `SecureStorage.*` | Зашифрованное хранилище для токенов/секретов; до 10 ключей на пользователя на бота (не web) |

Для JWT разумно хранить токен в `SecureStorage` с фолбэком на
`localStorage` в web.

## Биометрия (`BiometricManager`, не desktop/web)

`init()`, `isInited`, `isBiometricAvailable`, `isAccessRequested`,
`isAccessGranted`, `isBiometricTokenSaved`, `biometricType`,
`deviceId`, `requestAccess(reason?)`, `authenticate(reason?)`,
`updateBiometricToken(token?, reason?)`, `openSettings()`.
Типы: `finger` \| `face` \| `unknown` (Android всегда `unknown`).

## Хаптика (`HapticFeedback`, не desktop/web)

| Метод | Стиль |
|-------|-------|
| `impactOccurred(style, disableVibrationFallback?)` | `soft` \| `light` \| `medium` \| `heavy` \| `rigid` |
| `notificationOccurred(type, ...)` | `error` \| `success` \| `warning` |
| `selectionChanged(...)` | смена выбора (не подтверждение!) |

В наших тикетах: свайп в ленте → `impactOccurred('light')`,
матч → `notificationOccurred('success')`.

## NFC (`NfcManager`, только Android)

`init()`, `isInited`, `openSystemSettings()`, `emulateNfcTag(nfctag?)`.

## Ошибки

Методы возвращают Promise; reject:

```json
{ "error": { "code": "client.<метод>.<причина>" } }
```
