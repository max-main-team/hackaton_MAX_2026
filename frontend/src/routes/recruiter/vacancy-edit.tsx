import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { companyVacancies, myCompanies, updateVacancy } from '../../api/company'
import type { VacancyInput } from '../../api/types'
import { EMPLOYMENT_TYPES, WORK_FORMATS, WORK_FORMAT_LABELS } from '../../api/types'
import { digits } from '../../components/format'
import { TabBar } from '../../components/TabBar'

const TTL_OPTIONS = [12, 24, 48, 72, 168]

export default function VacancyEdit() {
  const navigate = useNavigate()
  const { id } = useParams()
  const vacancyId = Number(id)
  const [form, setForm] = useState<VacancyInput | null>(null)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  const load = useCallback(() => {
    myCompanies()
      .then((cs) => (cs[0] ? companyVacancies(cs[0].id) : []))
      .then((list) => {
        const vacancy = list.find((v) => v.id === vacancyId)
        if (!vacancy) {
          setError('Вакансия не найдена')
          return
        }
        setForm({
          title: vacancy.title,
          description: vacancy.description,
          required_skills: vacancy.required_skills,
          min_experience_months: vacancy.min_experience_months,
          city: vacancy.city,
          work_format: vacancy.work_format,
          employment_type: vacancy.employment_type,
          salary_min: vacancy.salary_min,
          salary_max: vacancy.salary_max,
          response_ttl_hours: vacancy.response_ttl_hours,
        })
      })
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [vacancyId])

  useEffect(load, [load])

  const set = <K extends keyof VacancyInput>(key: K, value: VacancyInput[K]) => {
    setForm((f) => (f ? { ...f, [key]: value } : f))
  }

  const save = async () => {
    setError('')
    if (!form) return
    if (!form.title.trim()) {
      setError('Укажите название вакансии')
      return
    }
    setSaving(true)
    try {
      await updateVacancy(vacancyId, form)
      navigate('/company/vacancies')
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  const skills = form?.required_skills
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean) ?? []

  return (
    <main className="screen">
      <button className="back" onClick={() => navigate('/company/vacancies')}>
        ← Вакансии
      </button>
      <header className="screen-header" style={{ paddingTop: 8 }}>
        <h1 style={{ fontSize: 20 }}>Редактирование</h1>
        <button className="link-btn" style={{ color: 'var(--muted)' }} onClick={() => navigate('/company/vacancies')}>
          Отмена
        </button>
      </header>
      <div className="screen-body has-tabbar" style={{ paddingBottom: 140, gap: 16 }}>
        {error && <p className="error-text">{error}</p>}
        {!form && !error && <p className="muted">Загрузка…</p>}

        {form && (
          <>
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
                    type="text" inputMode="numeric"
                    style={{ maxWidth: 90 }}
                    value={form.min_experience_months}
                    onChange={(e) => set('min_experience_months', Number(digits(e.target.value)))}
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
                  <span className="range">
                    <input
                      className="input plain"
                      type="text" inputMode="numeric"
                      placeholder="от"
                      style={{ width: `${Math.max(4, String(form.salary_min ?? '').length + 1)}ch` }}
                      value={form.salary_min ?? ''}
                      onChange={(e) => set('salary_min', digits(e.target.value) === '' ? null : Number(digits(e.target.value)))}
                    />
                    <span className="range-dash">–</span>
                    <input
                      className="input plain"
                      type="text" inputMode="numeric"
                      placeholder="до"
                      style={{ width: `${Math.max(4, String(form.salary_max ?? '').length + 1)}ch` }}
                      value={form.salary_max ?? ''}
                      onChange={(e) => set('salary_max', digits(e.target.value) === '' ? null : Number(digits(e.target.value)))}
                    />
                    <span className="fvalue">₽</span>
                  </span>
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
          </>
        )}
      </div>

      <div className="form-footer" style={{ bottom: 'calc(var(--tabbar-h) + env(safe-area-inset-bottom))' }}>
        <button className="btn btn-xl" onClick={save} disabled={saving || !form}>
          {saving ? 'Сохраняем…' : 'Сохранить'}
        </button>
      </div>

      <TabBar role="recruiter" />
    </main>
  )
}
