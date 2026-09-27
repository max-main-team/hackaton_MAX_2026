import { useEffect, useState } from 'react'
import { confirmActivity, getResume, saveResume } from '../../api/resume'
import type { Resume, ResumeInput, ResumeLink } from '../../api/types'
import { EMPLOYMENT_TYPES, WORK_FORMATS, WORK_FORMAT_LABELS } from '../../api/types'

const EMPTY_FORM: ResumeInput = {
  title: '',
  skills: '',
  experience_months: 0,
  about: '',
  education: '',
  links: [],
  city: '',
  work_format: 'onsite',
  employment_type: 'full_time',
  salary_min: null,
  salary_max: null,
}

function parseLinks(raw: string | ResumeLink[] | null | undefined): ResumeLink[] {
  if (Array.isArray(raw)) return raw
  try {
    const parsed = JSON.parse(raw ?? '[]')
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export default function ResumeScreen() {
  const [form, setForm] = useState<ResumeInput>(EMPTY_FORM)
  const [links, setLinks] = useState<ResumeLink[]>([])
  const [notFound, setNotFound] = useState(false)
  const [loading, setLoading] = useState(true)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState('')
  const [confirmed, setConfirmed] = useState(false)

  useEffect(() => {
    getResume()
      .then((r) => {
        setForm({
          title: r.title,
          skills: r.skills,
          experience_months: r.experience_months,
          about: r.about,
          education: r.education,
          links: [],
          city: r.city,
          work_format: r.work_format,
          employment_type: r.employment_type,
          salary_min: r.salary_min,
          salary_max: r.salary_max,
        })
        setLinks(parseLinks(r.links))
        setNotFound(false)
      })
      .catch((e) => {
        if (e instanceof Error && e.message.includes('not found')) setNotFound(true)
        else setError(String(e))
      })
      .finally(() => setLoading(false))
  }, [])

  const setField = <K extends keyof ResumeInput>(key: K, value: ResumeInput[K]) => {
    setForm((f) => ({ ...f, [key]: value }))
    setSaved(false)
  }

  const save = async () => {
    setError('')
    if (!form.title.trim()) {
      setError('Укажите желаемую должность')
      return
    }
    const cleanedLinks = links.filter((l) => l.type.trim() && l.url.trim())
    try {
      const savedResume: Resume = await saveResume({ ...form, links: cleanedLinks })
      setLinks(parseLinks(savedResume.links))
      setSaved(true)
      setNotFound(false)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }

  const confirm = async (active: boolean) => {
    try {
      await confirmActivity(active)
      setConfirmed(active)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }

  if (loading) {
    return (
      <main className="page">
        <p className="muted">Загрузка…</p>
      </main>
    )
  }

  return (
    <main className="page">
      <h1 className="page-title">Моё резюме</h1>
      {notFound && <p className="muted">Заполните резюме — оно появится в подборках компаний</p>}

      <div className="card">
        <div className="field">
          <label>Желаемая должность *</label>
          <input className="input" value={form.title} onChange={(e) => setField('title', e.target.value)} />
        </div>
        <div className="field">
          <label>Навыки (через запятую)</label>
          <input className="input" value={form.skills} onChange={(e) => setField('skills', e.target.value)} />
        </div>
        <div className="row">
          <div className="field" style={{ flex: 1 }}>
            <label>Опыт, мес</label>
            <input
              className="input"
              type="number"
              value={form.experience_months}
              onChange={(e) => setField('experience_months', Number(e.target.value))}
            />
          </div>
          <div className="field" style={{ flex: 1 }}>
            <label>Город</label>
            <input className="input" value={form.city} onChange={(e) => setField('city', e.target.value)} />
          </div>
        </div>
        <div className="row">
          <div className="field" style={{ flex: 1 }}>
            <label>Формат работы</label>
            <select className="select" value={form.work_format} onChange={(e) => setField('work_format', e.target.value)}>
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
              onChange={(e) => setField('employment_type', e.target.value)}
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
              onChange={(e) => setField('salary_min', e.target.value === '' ? null : Number(e.target.value))}
            />
          </div>
          <div className="field" style={{ flex: 1 }}>
            <label>до</label>
            <input
              className="input"
              type="number"
              value={form.salary_max ?? ''}
              onChange={(e) => setField('salary_max', e.target.value === '' ? null : Number(e.target.value))}
            />
          </div>
        </div>
        <div className="field">
          <label>О себе</label>
          <textarea className="textarea" value={form.about} onChange={(e) => setField('about', e.target.value)} />
        </div>
        <div className="field">
          <label>Образование</label>
          <input className="input" value={form.education} onChange={(e) => setField('education', e.target.value)} />
        </div>

        <div className="field">
          <label>Ссылки</label>
          {links.map((link, i) => (
            <div className="row" key={i}>
              <input
                className="input"
                style={{ maxWidth: 110 }}
                value={link.type}
                placeholder="github"
                onChange={(e) =>
                  setLinks(links.map((x, xi) => (xi === i ? { ...x, type: e.target.value } : x)))
                }
              />
              <input
                className="input"
                style={{ flex: 1 }}
                value={link.url}
                placeholder="https://..."
                onChange={(e) =>
                  setLinks(links.map((x, xi) => (xi === i ? { ...x, url: e.target.value } : x)))
                }
              />
              <button
                className="btn btn-secondary"
                style={{ padding: '8px 10px' }}
                onClick={() => setLinks(links.filter((_, xi) => xi !== i))}
              >
                ✕
              </button>
            </div>
          ))}
          <button className="link-btn" onClick={() => setLinks([...links, { type: '', url: '' }])}>
            + добавить ссылку
          </button>
        </div>

        {saved && <p className="ok-text">Сохранено</p>}
        {error && <p className="error-text">{error}</p>}
        <button className="btn" onClick={save}>
          Сохранить резюме
        </button>
      </div>

      <div className="card">
        <strong>Подбор актуален?</strong>
        <span className="muted">Подтверждайте раз в неделю, чтобы компании вас видели</span>
        {confirmed && <span className="ok-text">Спасибо, подтвердили</span>}
        <div className="row">
          <button className="btn btn-success" onClick={() => confirm(true)}>
            Актуально
          </button>
          <button className="btn btn-secondary" onClick={() => confirm(false)}>
            Не ищу работу
          </button>
        </div>
      </div>
    </main>
  )
}
