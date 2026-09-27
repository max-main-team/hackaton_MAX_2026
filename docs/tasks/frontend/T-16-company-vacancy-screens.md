# T-16 — Рекрутер: компания и вакансии

- **Статус:** ✅
- **Спека:** `docs/frontend/spec-ui.md` (разделы «Компания», «Настройка TTL», «Вакансии»)
- **Зависимости:** T-14
- **Разблокирует:** T-17

## Что сделать

- `routes/recruiter/company.tsx`: `GET /my/companies` пуст → форма
  создания (name, description, website, logo_url, address + селект своей
  позиции `owner | hr | employee` — создатель не обязан быть owner);
  иначе карточка компании (с `Badge` verified) + переход к вакансиям.
- `routes/recruiter/vacancy-new.tsx`: поля вакансии по контракту
  `VacancyInput` — title, description, required_skills,
  min_experience_months, city, work_format, employment_type,
  salary_min/salary_max (опционально) + TTL-селект
  (12/24/48/72/168 ч, дефолт 48) → `POST /companies/{id}/vacancies`
  → переход к списку.
- `routes/recruiter/vacancies.tsx`: список вакансий компании
  (title, город, TTL, active/черновик), кнопки: «Лента» (в T-17),
  «Список» (в T-17), переключить active (`PATCH`).
- `components/Badge.tsx` — бейдж «верифицирована».

## Критерии приёмки

- [ ] Цепочка: создать компанию → создать вакансию → видна в списке.
- [ ] TTL из селекта доезжает до бека (проверить ответом GET).
- [ ] `npm run build && npm run lint` — чисто.
