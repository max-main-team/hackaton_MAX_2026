import { useEffect, useRef, useState } from 'react'
import { confirmActivity, getResume, parseFile, parseResume, saveResume } from '../../api/resume'
import type { Resume, ResumeInput, ResumeLink } from '../../api/types'
import { EMPLOYMENT_TYPES, WORK_FORMATS, WORK_FORMAT_LABELS } from '../../api/types'
import { Screen } from '../../components/Screen'
import { Icon } from '../../components/Icon'
import { ApiError } from '../../api/client'
import { digits } from '../../components/format'
import { authHeaders } from '../../lib/session'

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

const MAX_PDF_BYTES = 10 * 1024 * 1024

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
  const [stage, setStage] = useState<'idle' | 'extracting' | 'parsing'>('idle')
  const [elapsed, setElapsed] = useState(0)
  const [experienceDraft, setExperienceDraft] = useState<string | null>(null)
  const [confirmed, setConfirmed] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (stage !== 'parsing' && stage !== 'extracting') return
    setElapsed(0)
    const timer = setInterval(() => setElapsed(s => s + 1), 1000)
    return () => clearInterval(timer)
  }, [stage])

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
        setExperienceDraft(null)
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

  const setExperience = (raw: string) => {
    const clean = digits(raw)
    setExperienceDraft(clean)
    setField('experience_months', clean === '' ? 0 : Number(clean))
  }

  const confirm = async (active: boolean) => {
    try {
      await confirmActivity(active)
      setConfirmed(active)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }

  const validate = (): string | null => {
    if (!form.title.trim()) return 'Укажите желаемую должность'
    return null
  }

  const save = async () => {
    const problem = validate()
    if (problem) {
      setError(problem)
      return
    }
    setError('')
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

  const uploadPdf = async (file: File) => {
    setError('')
    const isPdf = file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf')
    if (!isPdf) {
      setError('Загрузите файл в формате PDF')
      return
    }
    if (file.size > MAX_PDF_BYTES) {
      setError('Файл больше 10 МБ')
      return
    }

    setStage('extracting')

    // 1) сервер-сайд: файл уходит на бекенд, текст извлекается там
    let serverAccepted = false
    try {
      await parseFile(file)
      serverAccepted = true
    } catch (e) {
      const detail = e instanceof Error ? e.message : String(e)
      void fetch('/api/v1/debug/log', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify({ where: 'resume-parse-file', detail }),
      }).catch(() => {})
      if (e instanceof ApiError && e.status === 400 && /larger than 10/i.test(e.message)) {
        setError('Файл больше 10 МБ')
        return
      }
      // остальное (нет текстового слоя и т.п.) → пробуем в браузере
    }

    // 2) фолбэк: клиентский pdfjs (legacy)
    if (!serverAccepted) {
      let text = ''
      try {
        const { extractPdfText } = await import('../../lib/pdf')
        text = await extractPdfText(file)
      } catch (e) {
        const detail = e instanceof Error ? `${e.name}: ${e.message}` : String(e)
        void fetch('/api/v1/debug/log', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', ...authHeaders() },
          body: JSON.stringify({ where: 'resume-extract', detail }),
        }).catch(() => {})
        setError('Не удалось обработать PDF — заполните резюме вручную или попробуйте другой файл')
        return
      }
      try {
        await parseResume(text, file.name)
      } catch (e) {
        const detail = e instanceof Error ? e.message : String(e)
        void fetch('/api/v1/debug/log', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', ...authHeaders() },
          body: JSON.stringify({ where: 'resume-parse', detail }),
        }).catch(() => {})
        setError('Не удалось обработать PDF — заполните резюме вручную или попробуйте другой файл')
        return
      }
    }

    // 3) AI в фоне — поллинг статуса до 6 минут
    for (let i = 0; i < 90; i++) {
      await new Promise(r => setTimeout(r, 4000))
      const r = await getResume()
      if (r.parse_status === 'failed') {
        setError('Не удалось распознать PDF — заполните резюме вручную')
        return
      }
      if (r.parse_status === 'done') {
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
        setSaved(false)
        setExperienceDraft(null)
        return
      }
    }
    setError('Обработка заняла слишком много времени — резюме заполнится чуть позже, обновите экран')
  }

  if (loading) {
    return (
      <Screen role="candidate">
        <p className="muted">Загрузка…</p>
      </Screen>
    )
  }

  return (
    <Screen role="candidate" title="Моё резюме" icon="user" sub={notFound ? 'Заполните резюме или загрузите PDF — оно появится в подборках компаний' : undefined}>
      {error && <p className="error-text">{error}</p>}

      <div className="ai-block">
        <span className="ai-label">
          <Icon name="zap" size={14} />
          Быстрый старт
        </span>
        <p style={{ fontSize: 13 }}>
          Загрузите PDF-резюме — AI заполнит поля. Проверьте и поправьте перед сохранением.
        </p>
        <input
          ref={fileRef}
          type="file"
          accept=".pdf,application/pdf"
          onChange={(e) => {
            const file = e.target.files?.[0]
            if (file) void uploadPdf(file)
          }}
          disabled={stage !== 'idle'}
          style={{ fontSize: 12, color: 'var(--muted)' }}
        />
        {stage === 'extracting' && (
          <p className="muted" style={{ margin: 0 }}>
            Извлекаем текст из PDF… {elapsed} сек
          </p>
        )}
        {stage === 'parsing' && (
          <p className="muted" style={{ margin: 0 }}>
            AI распознаёт резюме… {elapsed} сек. Большое резюме занимает до 2–3 минут — не закрывайте экран
          </p>
        )}
      </div>

      <div className="form-section">
        <p className="section-label">Основное</p>
        <div className="fieldset">
          <div className="field">
            <label>Желаемая должность *</label>
            <input className="input" value={form.title} onChange={(e) => setField('title', e.target.value)} />
          </div>
          <div className="field">
            <label>Навыки (через запятую)</label>
            <input className="input" value={form.skills} onChange={(e) => setField('skills', e.target.value)} />
          </div>
        </div>
      </div>

      <div className="form-section">
        <p className="section-label">Условия</p>
        <div className="fieldset">
          <div className="fieldset-row">
            <span className="flabel">Опыт, мес</span>
            <input
              className="input plain"
              type="text" inputMode="numeric"
              placeholder="0"
              style={{ maxWidth: 90 }}
              value={experienceDraft ?? (form.experience_months === 0 ? '' : String(form.experience_months))}
              onChange={(e) => setExperience(e.target.value)}
            />
          </div>
          <hr className="divider" />
          <div className="fieldset-row">
            <span className="flabel">Город</span>
            <input
              className="input plain"
              style={{ maxWidth: 180 }}
              value={form.city}
              onChange={(e) => setField('city', e.target.value)}
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
                  onClick={() => setField('work_format', f)}
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
              onChange={(e) => setField('employment_type', e.target.value)}
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
                onChange={(e) => setField('salary_min', digits(e.target.value) === '' ? null : Number(digits(e.target.value)))}
              />
              <span className="range-dash">–</span>
              <input
                className="input plain"
                type="text" inputMode="numeric"
                placeholder="до"
                style={{ width: `${Math.max(4, String(form.salary_max ?? '').length + 1)}ch` }}
                value={form.salary_max ?? ''}
                onChange={(e) => setField('salary_max', digits(e.target.value) === '' ? null : Number(digits(e.target.value)))}
              />
              <span className="fvalue">₽</span>
            </span>
          </div>
        </div>
      </div>

      <div className="form-section" style={{ paddingBottom: 8 }}>
        <p className="section-label">Подбор актуален?</p>
        <div className="card gap-sm">
          <span className="muted" style={{ fontSize: 12 }}>
            Подтверждайте раз в неделю, чтобы компании вас видели
          </span>
          {confirmed && <span className="ok-text">Спасибо, подтвердили</span>}
          <div className="row">
            <button className="btn" style={{ flex: 1 }} onClick={() => confirm(true)}>
              Актуально
            </button>
            <button className="btn btn-ghost" style={{ flex: 1 }} onClick={() => confirm(false)}>
              Не ищу работу
            </button>
          </div>
        </div>
      </div>

      <div className="form-section">
        <p className="section-label">О себе</p>
        <div className="fieldset">
          <textarea
            className="textarea"
            style={{ background: 'transparent', border: 'none', padding: 0, minHeight: 64 }}
            placeholder="Расскажите о своём опыте…"
            value={form.about}
            onChange={(e) => setField('about', e.target.value)}
          />
          <hr className="divider" />
          <div className="field">
            <label>Образование</label>
            <input className="input" value={form.education} onChange={(e) => setField('education', e.target.value)} />
          </div>
        </div>
      </div>

      <div className="form-section">
        <p className="section-label">Ссылки</p>
        <div className="fieldset">
          {links.map((link, i) => (
            <div className="row" key={i} style={{ flexWrap: 'nowrap' }}>
              <input
                className="input"
                style={{ maxWidth: 110 }}
                value={link.type}
                placeholder="github"
                onChange={(e) => setLinks(links.map((x, xi) => (xi === i ? { ...x, type: e.target.value } : x)))}
              />
              <input
                className="input"
                style={{ flex: 1 }}
                value={link.url}
                placeholder="https://..."
                onChange={(e) => setLinks(links.map((x, xi) => (xi === i ? { ...x, url: e.target.value } : x)))}
              />
              <button
                className="btn btn-ghost"
                style={{ width: 38, padding: 0, flex: 'none' }}
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
      </div>

      {saved && <p className="ok-text" style={{ margin: 0 }}>Сохранено</p>}

      <button className="btn btn-xl" onClick={save}>
        Сохранить резюме
      </button>


    </Screen>
  )
}
