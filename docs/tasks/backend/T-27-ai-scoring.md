# T-27 — AI-скоринг (LLM-оценка соответствия)

- **Статус:** ⬜ (обвязка + рабочий AI при наличии рабочего ключа)
- **Спека:** docs/common/features.md (фича 1.3, AI-часть)
- **Зависимости:** T-08

## Контекст

Ключ от z.ai (coding plan) предоставлен владельцем. Точный режим работы
ключа против прямого API неизвестен — поэтому делаем **универсальную
обвязку** под любой OpenAI-совместимый провайдер + фолбэк.

## Env (в .env.example и GH-секреты; ключ в репо НЕ хранить!)

| Переменная    | Дефолт                                    |
|---------------|-------------------------------------------|
| `AI_API_KEY`  | пусто → AI выключен, только алго-скоринг  |
| `AI_BASE_URL` | `https://api.z.ai/api/paas/v4`            |
| `AI_MODEL`    | `glm-4.6`                                 |

Возможные кандидаты base_url для z.ai: `https://api.z.ai/api/paas/v4`,
`https://open.bigmodel.cn/api/paas/v4`. Проверять `GET {base}/chat/completions`
POST-запросом; если ключ «кодинговый» и не работает на прямом API —
фолбэк на алго-скоринг, обвязка остаётся.

## Что сделать

1. **Миграция 000006**: `resume_scores (id, resume_id FK, vacancy_id FK,
   algo_score INT, ai_score INT NULL, ai_comment TEXT NULL, ai_model TEXT,
   created_at, updated_at, UNIQUE (resume_id, vacancy_id))`.
2. `internal/scoring/ai.go`:
   - `ScoreAI(v Vacancy, r Resume) (int, string, error)` — POST
     `{AI_BASE_URL}/chat/completions`, модель `AI_MODEL`, prompt: JSON
     резюме+вакансии → строго JSON-ответ `{"score":0-100,"comment":"..."}`.
   - Таймаут 15с; любая ошибка → фолбэк на алго, сообщение в лог.
3. Кеш в `resume_scores`: сначала читаем кеш (свежий), при изменении
   резюме/вакансии — пересчёт.
4. Блендинг в выдаче `GET /vacancies/{id}/candidates`:
   `final = round(0.6*algo + 0.4*ai)` при рабочем AI, иначе `algo`.
   В ответ добавляются `ai_comment` и `final_score`.
5. При старте, если `AI_API_KEY` задан — один проверочный запрос,
   результат (ok/fail + модель) в лог.

## Критерии приёмки

- [ ] Без `AI_API_KEY` всё работает как раньше (только algo).
- [ ] С рабочим ключом: `/candidates` возвращает `final_score` + `ai_comment`,
      результат закеширован в `resume_scores`.
- [ ] Падение LLM-ручки не ломает выдачу (фолбэк).
- [ ] go build/vet/gofmt и тесты — чисто.
