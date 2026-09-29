# BUG-06 — PDF-загрузка не работает в вебвью MAX (iOS/Android)

- **Приоритет:** 🔴 блокер фичи T-30 в MAX
- **Экран:** кандидат → «Моё резюме» → загрузка PDF
- **Статус:** ✅

## Симптомы

- В браузере (десктоп Chrome) загрузка PDF работает.
- В MAX на iPhone: `undefined is not a function (near '...e of t...')`
  (формат ошибки JSC), после деплоя фикса —
  `Setting up fake worker failed: Failed to fetch dynamically imported
  module: …pdf.worker.min.mjs`.

## Причина

`pdfjs-dist@6.3.289` (включая legacy-сборку) использует:
- `Promise.withResolvers` — Safari 17.4+, в вебвью MAX отсутствует
  (`for...of` внутри pdfjs получает undefined-итератор → «near '...e of
  t...'»);
- `Object.fromEntries`, `structuredClone` — тоже отсутствуют в старых
  вебвью (вскрылись по цепочке после полифилла withResolvers);
- реальный модульный Worker в вебвью MAX не стартует, а fallback pdfjs
  («fake worker») грузит воркер через динамический `import()` — тоже
  падает: `Failed to fetch dynamically imported module`.

Воспроизведено локально: Playwright + addInitScript, удаляющий эти API.

## Фикс

- `frontend/src/polyfills.ts` (подключён раньше всех в `main.tsx`):
  полифиллы `Promise.withResolvers`, `Object.fromEntries`,
  `Object.hasOwn`, `Array.prototype.at/findLast` (non-enumerable —
  pdfjs проверяет целостность Array.prototype),
  `String.prototype.replaceAll`, `globalThis.structuredClone`.
- `frontend/src/lib/pdf.ts`: воркер отключён — legacy-модуль воркера
  импортируется статически и кладётся в `globalThis.pdfjsWorker`, pdfjs
  парсит PDF на main-треде (вебвью перестаёт зависеть от Worker и
  динамических импортов).
- Плюс валидация типа файла в `uploadPdf`: только PDF, иначе ошибка
  «Загрузите файл в формате PDF» (вебвью MAX игнорирует `accept`).

## Проверка

- [x] Локально с эмуляцией старого вебвью (снос withResolvers/
      fromEntries/structuredClone/at/replaceAll/findLast): PDF
      извлекается, `POST /my/resume/parse` уходит (200).
- [x] Фейковый токен файла (не-PDF) → «Загрузите файл в формате PDF».
- [x] Прод: AI-заполнение полей работает (AI настроен на проде).
