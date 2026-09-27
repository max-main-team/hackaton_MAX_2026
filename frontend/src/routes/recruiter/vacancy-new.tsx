import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { createVacancy, myCompanies } from '../../api/company'
import type { VacancyInput } from '../../api/types'
import { EMPLOYMENT_TYPES, WORK_FORMATS, WORK_FORMAT_LABELS } from '../../api/types'

const TTL_OPTIONS = [12, 24, 48, 72, 168]

export default function VacancyNew() {
  const navigate = useNavigate()
  const [companyId, setCompanyId] = useState<number | null>(null)
  const [form, setForm] = useState<VacancyInput>({
    title: '',
    description: '',
    required_skills: '',
    min_experience_months: 0,
    city: 'Санкт-Петербург',
    work_format: 'hybrid',
    employment_type: 'full_time',
    salary_min: null,
    salary_max: null,
    response_ttl_hours: 48,
  })
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

  return (
    <main className="page">
      <Link className="back" to="/company/vacancies">
        ← Вакансии
      </Link>
      <h1 className="page-title">Новая вакансия</h1>

      <div className="card">
        <div className="field">
          <label>Название *</label>
          <input className="input" value={form.title} onChange={(e) => set('title', e.target.value)} />
        </div>
        <div className="field">
          <label>Описание</label>
          <textarea className="textarea" value={form.description} onChange={(e) => set('description', e.target.value)} />
        </div>
        <div className="field">
          <label>Требуемые навыки (через запятую)</label>
          <input className="input" value={form.required_skills} onChange={(e) => set('required_skills', e.target.value)} />
        </div>
        <div className="row">
          <div className="field" style={{ flex: 1 }}>
            <label>Мин. опыт, мес</label>
            <input
              className="input"
              type="number"
              value={form.min_experience_months}
              onChange={(e) => set('min_experience_months', Number(e.target.value))}
            />
          </div>
          <div className="field" style={{ flex: 1 }}>
            <label>Город</label>
            <input className="input" value={form.city} onChange={(e) => set('city', e.target.value)} />
          </div>
        </div>
        <div className="row">
          <div className="field" style={{ flex: 1 }}>
            <label>Формат работы</label>
            <select className="select" value={form.work_format} onChange={(e) => set('work_format', e.target.value)}>
              {WORK_FORMATS.map((f) => (
                <option key={f} value={f}>
                  {WORK_FORMAT_LABELS[f] ?? f}
                </option>
              ))}
            </select>
          </div>
          <div className="field" style={{ flex: 1 }}>
            <label>Занятость</label>
            <select
              className="select"
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
        </div>
        <div className="row">
          <div className="field" style={{ flex: 1 }}>
            <label>Зарплата от</label>
            <input
              className="input"
              type="number"
              value={form.salary_min ?? ''}
              onChange={(e) => set('salary_min', e.target.value === '' ? null : Number(e.target.value))}
            />
          </div>
          <div className="field" style={{ flex: 1 }}>
            <label>до</label>
            <input
              className="input"
              type="number"
              value={form.salary_max ?? ''}
              onChange={(e) => set('salary_max', e.target.value === '' ? null : Number(e.target.value))}
            />
          </div>
        </div>
        <div className="field">
          <label>⏱ Срок ответа рекрутера (TTL)</label>
          <select
            className="select"
            value={form.response_ttl_hours}
            onChange={(e) => set('response_ttl_hours', Number(e.target.value))}
          >
            {TTL_OPTIONS.map((h) => (
              <option key={h} value={h}>
                {h} ч
              </option>
            ))}
          </select>
        </div>

        {error && <p className="error-text">{error}</p>}
        <button className="btn" onClick={save} disabled={saving || companyId === null}>
          {saving ? 'Сохраняем…' : 'Создать вакансию'}
        </button>
      </div>
    </main>
  )
}
