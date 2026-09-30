import { useEffect, useRef, useState } from 'react'
import { allCandidates } from '../../api/matching'
import type { CandidateCard } from '../../api/types'
import { WORK_FORMAT_LABELS } from '../../api/types'
import { TabBar } from '../../components/TabBar'
import { experienceLabel, formatSalary } from '../../components/format'

export default function AllCandidates() {
  const [items, setItems] = useState<CandidateCard[]>([])
  const [total, setTotal] = useState(0)
  const [query, setQuery] = useState('')
  const [expanded, setExpanded] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const load = (q: string, off: number) => {
    allCandidates(q, 20, off)
      .then((d) => {
        setItems((prev) => (off === 0 ? d.items : [...prev, ...d.items]))
        setTotal(d.total)
      })
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    if (timer.current) clearTimeout(timer.current)
    timer.current = setTimeout(() => load(query, 0), 300)
    return () => {
      if (timer.current) clearTimeout(timer.current)
    }
  }, [query])

  return (
    <main className="screen">
      <header className="screen-header" style={{ paddingTop: 8 }}>
        <span className="chip" style={{ padding: '6px 10px', borderRadius: 6, fontSize: 13, fontWeight: 700, color: 'var(--text)' }}>
          Все кандидаты ({total})
        </span>
      </header>
      <div className="screen-body has-tabbar" style={{ gap: 12 }}>
        <input
          className="input"
          placeholder="Поиск: должность, навыки, город"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        {error && <p className="error-text">{error}</p>}
        {loading && <p className="muted">Загрузка…</p>}
        {!loading && items.length === 0 && (
          <div className="center-note">
            <span>Пока никто не найден</span>
          </div>
        )}
        {items.map((card) => (
          <div className="list-item" key={card.user.id}>
            <div
              className="row between"
              style={{ cursor: 'pointer', flexWrap: 'nowrap' }}
              onClick={() => setExpanded(expanded === card.user.id ? null : card.user.id)}
            >
              <strong style={{ fontSize: 14, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {card.user.first_name} {card.user.last_name}
              </strong>
              <span className="muted" style={{ fontSize: 12, whiteSpace: 'nowrap' }}>
                {experienceLabel(card.resume.experience_months)}
              </span>
            </div>
            <span className="muted" style={{ fontSize: 12 }}>
              {card.resume.title} · {card.resume.city}
            </span>
            {expanded === card.user.id && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8, paddingTop: 8 }}>
                <span className="muted" style={{ fontSize: 12 }}>
                  {WORK_FORMAT_LABELS[card.resume.work_format] ?? card.resume.work_format} ·{' '}
                  {WORK_FORMAT_LABELS[card.resume.employment_type] ?? card.resume.employment_type}
                  {card.resume.salary_min || card.resume.salary_max
                    ? ` · ${formatSalary(card.resume.salary_min, card.resume.salary_max)}`
                    : ''}
                </span>
                {card.resume.skills && (
                  <div className="stack-tags">
                    {card.resume.skills
                      .split(',')
                      .map((s) => s.trim())
                      .filter(Boolean)
                      .slice(0, 12)
                      .map((s) => (
                        <span className="chip" key={s}>
                          {s}
                        </span>
                      ))}
                  </div>
                )}
                {card.resume.about && <p style={{ fontSize: 13 }}>{card.resume.about}</p>}
                {card.resume.education && (
                  <span className="muted" style={{ fontSize: 12 }}>
                    {card.resume.education}
                  </span>
                )}
                <span className="muted" style={{ fontSize: 11 }}>
                  Приглашение можно отправить из подбора по вакансии
                </span>
              </div>
            )}
          </div>
        ))}
        {items.length < total && (
          <button className="btn btn-ghost" onClick={() => load(query, items.length)}>
            Показать ещё
          </button>
        )}
      </div>
      <TabBar role="recruiter" />
    </main>
  )
}
