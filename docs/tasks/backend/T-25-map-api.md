# T-25 — Карта вакансий (бек)

- **Статус:** ⬜
- **Спека:** docs/common/features.md (фича 4)
- **Зависимости:** T-07

## Что сделать

- Миграция: `ALTER TABLE vacancies ADD COLUMN lat DOUBLE PRECISION, ADD COLUMN lng DOUBLE PRECISION`.
- Словарь городов (бек, internal/geo/geo.go): Санкт-Петербург (59.9386, 30.3141),
  Москва (55.7558, 37.6173), Новгород, Казань, Екатеринбург — при создании
  вакансии координаты проставляются по `city` (неизвестный город → NULL).
- `GET /vacancies/map` — активные вакансии с координатами:
  `[{id, title, company_name, city, lat, lng, salary_min, salary_max}]`.

## Критерии приёмки

- [ ] Вакансия в СПб получает координаты СПб.
- [ ] Неизвестный город → lat/lng NULL, в /map не попадает.
- [ ] go build/vet/gofmt — чисто.
