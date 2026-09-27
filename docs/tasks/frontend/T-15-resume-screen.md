# T-15 — Экран резюме кандидата

- **Статус:** ✅
- **Спека:** `docs/frontend/spec-ui.md` (раздел «Резюме»), контракт `docs/backend/spec-api.md`
- **Зависимости:** T-14
- **Разблокирует:** T-17 (лента видит резюме)

## Что сделать

Форма резюме — поля по контракту `PUT /my/resume` (см. Swagger `ResumeInput`):

- `title` (обязательное), `skills` (строка через запятую или теги),
- `experience_months` (0..600), `about`, `education`,
- `links` — динамический список пар {type, url} (type: github/linkedin/habr/other),
  сериализуется в JSON-массив,
- `city`,
- `work_format` — селект `onsite | hybrid | remote` (ОБЯЗАТЕЛЬНО, вместо
  старого schedule),
- `employment_type` — селект `full_time | part_time | contract | internship`,
- `salary_min`, `salary_max` — опциональные числа.

Загрузка: `GET /my/resume` → 404 = пустая форма, иначе предзаполнение
(включая `links` — это уже JSON-массив с бека).

Сохранение: `PUT /my/resume`; состояния loading / error / saved
(короткий тост или подпись «Сохранено»).

Валидация на клиенте дублирует бек: title непустой, `work_format` и
`employment_type` из списков, experience 0..600, salary ≥ 0.

## Критерии приёмки

- [ ] Создание и повторное редактирование работают, links сохраняются и
      возвращаются массивом.
- [ ] Невалидная форма не отправляется, ошибка бека показывается.
- [ ] `npm run build && npm run lint` — чисто.
