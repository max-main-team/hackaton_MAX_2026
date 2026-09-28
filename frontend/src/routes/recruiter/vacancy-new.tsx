import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createVacancy, myCompanies } from '../../api/company'
import type { VacancyInput } from '../../api/types'
import { EMPLOYMENT_TYPES, WORK_FORMATS, WORK_FORMAT_LABELS } from '../../api/types'
import { TabBar } from '../../components/TabBar'

const TTL_OPTIONS = [12, 24, 48, 72, 168]

const EMPTY_FORM: VacancyInput = {
  title: '',
  description: '',
  required_skills: '',
  min_experience_months: 0,
  city: 'Санкт-Петербург',
  work_format: 'remote',
  employment_type: 'full_time',
  salary_min: null,
  salary_max: null,
  response_ttl_hours: 48,
}

export default function VacancyNew() {
  const navigate = useNavigate()
  const [companyId, setCompanyId] = useState<number | null>(null)
  const [form, setForm] = useState<VacancyInput>(EMPTY_FORM)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    myCompanies()
      .then((cs) => {
        if (cs[0]) setCompanyId(cs[0].id)
        else setError('Сначала создайте компанию')
      })
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  const set = <K extends keyof VacancyInput>(key: K, value: VacancyInput[K]) => {
    setForm((f) => ({ ...f, [key]: value }))
  }

  const reset = () => setForm(EMPTY_FORM)

  const save = async () => {
    setError('')
    if (companyId === null) return
    if (!form.title.trim()) {
      setError('Укажите название вакансии')
      return
    }
    setSaving(true)
    try {
      await createVacancy(companyId, form)
      navigate('/company/vacancies')
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  const skills = form.required_skills
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)

  return (
    <main className="screen">
      <button className="back" onClick={() => navigate('/company/vacancies')}>
        ← Вакансии
      </button>
      <header className="screen-header" style={{ paddingTop: 8 }}>
        <h1 style={{ fontSize: 20 }}>Новая вакансия</h1>
        <button className="link-btn" style={{ color: 'var(--muted)' }} onClick={reset}>
          Сбросить
        </button>
      </header>
      <div className="screen-body has-tabbar" style={{ paddingBottom: 140, gap: 16 }}>
        {error && <p className="error-text">{error}</p>}

        <div className="form-section">
          <p className="section-label">Основное</p>
          <div className="fieldset">
            <input
              className="input plain bold"
              style={{ textAlign: 'left' }}
              placeholder="Название вакансии *"
              value={form.title}
              onChange={(e) => set('title', e.target.value)}
            />
            <hr className="divider" />
            <textarea
              className="textarea plain"
              style={{ textAlign: 'left', minHeight: 48 }}
              placeholder="Описание вакансии…"
              value={form.description}
              onChange={(e) => set('description', e.target.value)}
            />
            <hr className="divider" />
            <input
              className="input plain"
              style={{ textAlign: 'left', color: 'var(--muted)' }}
              placeholder="Навыки через запятую…"
              value={form.required_skills}
              onChange={(e) => set('required_skills', e.target.value)}
            />
            {skills.length > 0 && (
              <div className="row" style={{ gap: 6 }}>
                {skills.map((s) => (
                  <span className="chip accent" key={s}>
                    {s}
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="form-section">
          <p className="section-label">Условия</p>
          <div className="fieldset">
            <div className="fieldset-row">
              <span className="flabel">Опыт работы</span>
              <input
                className="input plain"
                type="number"
                style={{ maxWidth: 90 }}
                value={form.min_experience_months}
                onChange={(e) => set('min_experience_months', Number(e.target.value))}
              />
            </div>
            <hr className="divider" />
            <div className="fieldset-row">
              <span className="flabel">Город</span>
              <input
                className="input plain"
                style={{ maxWidth: 180 }}
                value={form.city}
                onChange={(e) => set('city', e.target.value)}
              />
            </div>
            <hr className="divider" />
            <div className="fieldset-row">
              <span className="flabel">Формат</span>
              <div className="seg">
                {WORK_FORMATS.map((f) => (
                  <button
                    key={f}
                    type="button"
                    className={`seg-item${form.work_format === f ? ' active' : ''}`}
                    onClick={() => set('work_format', f)}
                  >
                    {WORK_FORMAT_LABELS[f] ?? f}
                  </button>
                ))}
              </div>
            </div>
            <hr className="divider" />
            <div className="fieldset-row">
              <span className="flabel">Занятость</span>
              <select
                className="input plain"
                style={{ maxWidth: 150, appearance: 'none', cursor: 'pointer' }}
                value={form.employment_type}
                onChange={(e) => set('employment_type', e.target.value)}
              >
                {EMPLOYMENT_TYPES.map((t) => (
                  <option key={t} value={t}>
                    {WORK_FORMAT_LABELS[t] ?? t}
                  </option>
                ))}
              </select>
            </div>
            <hr className="divider" />
            <div className="fieldset-row">
              <span className="flabel">Оклад</span>
              <div className="row" style={{ flexWrap: 'nowrap', gap: 4 }}>
                <input
                  className="input plain"
                  type="number"
                  style={{ maxWidth: 90 }}
                  placeholder="от"
                  value={form.salary_min ?? ''}
                  onChange={(e) => set('salary_min', e.target.value === '' ? null : Number(e.target.value))}
                />
                <span className="muted">–</span>
                <input
                  className="input plain"
                  type="number"
                  style={{ maxWidth: 90 }}
                  placeholder="до"
                  value={form.salary_max ?? ''}
                  onChange={(e) => set('salary_max', e.target.value === '' ? null : Number(e.target.value))}
                />
                <span className="fvalue">₽</span>
              </div>
            </div>
          </div>
        </div>

        <div className="form-section">
          <p className="section-label">Срок ответа</p>
          <div className="ttl-row">
            {TTL_OPTIONS.map((h) => (
              <button
                key={h}
                type="button"
                className={`ttl-chip${form.response_ttl_hours === h ? ' active' : ''}`}
                onClick={() => set('response_ttl_hours', h)}
              >
                {h}ч
              </button>
            ))}
          </div>
          <p className="muted" style={{ fontSize: 11, margin: 0 }}>
            Кандидат увидит дедлайн. После истечения приглашение отменяется.
          </p>
        </div>
      </div>

      <div className="form-footer" style={{ bottom: 'calc(var(--tabbar-h) + env(safe-area-inset-bottom))' }}>
        <button className="btn btn-xl" onClick={save} disabled={saving || companyId === null}>
          {saving ? 'Публикуем…' : 'Опубликовать'}
        </button>
      </div>

      <TabBar role="recruiter" />
    </main>
  )
}
