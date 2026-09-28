import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { candidateAction, candidates } from '../../api/matching'
import type { CandidateItem } from '../../api/types'
import { WORK_FORMAT_LABELS } from '../../api/types'
import { TabBar } from '../../components/TabBar'
import { UserCard } from '../../components/UserCard'

export default function CandidateList() {
  const { id } = useParams()
  const vacancyId = Number(id)
  const [items, setItems] = useState<CandidateItem[]>([])
  const [total, setTotal] = useState(0)
  const [expanded, setExpanded] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const load = (off: number) => {
    candidates(vacancyId, 'list', 20, off)
      .then((d) => {
        setItems((prev) => (off === 0 ? d.items : [...prev, ...d.items]))
        setTotal(d.total)
      })
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
      .finally(() => setLoading(false))
  }

  useEffect(() => load(0), [vacancyId])

  const act = async (candidateUserId: number, action: 'invite' | 'skip') => {
    try {
      await candidateAction(vacancyId, candidateUserId, action)
      setItems((list) => list.filter((x) => x.user.id !== candidateUserId))
      setTotal((t) => Math.max(0, t - 1))
      setExpanded(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }

  return (
    <main className="screen">
      <button className="back" onClick={() => window.history.back()}>
        ← Вакансии
      </button>
      <header className="screen-header" style={{ paddingTop: 8 }}>
        <span className="chip" style={{ padding: '6px 10px', borderRadius: 6, fontSize: 13, fontWeight: 700, color: 'var(--text)' }}>
          Кандидаты ({total})
        </span>
        <div className="toggle-group">
          <Link className="toggle-chip" to={`/vacancy/${vacancyId}/feed`}>
            Карточки
          </Link>
          <span className="toggle-chip active">Список</span>
        </div>
      </header>
      <div className="screen-body has-tabbar" style={{ gap: 12 }}>
        {error && <p className="error-text">{error}</p>}
        {loading && <p className="muted">Загрузка…</p>}

        {!loading && items.length === 0 && (
          <div className="center-note">
            <span>Все кандидаты разобраны</span>
          </div>
        )}

        {items.map((item) => (
          <div className="list-item" key={item.user.id}>
            <div
              className="row between"
              style={{ cursor: 'pointer', flexWrap: 'nowrap' }}
              onClick={() => setExpanded(expanded === item.user.id ? null : item.user.id)}
            >
              <strong style={{ fontSize: 14, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {item.user.first_name} {item.user.last_name}
              </strong>
              <span className="chip accent">{Math.round(item.final_score)} Match</span>
            </div>
            <span className="muted" style={{ fontSize: 12 }}>
              {item.resume.title} · {item.resume.city} ·{' '}
              {WORK_FORMAT_LABELS[item.resume.work_format] ?? item.resume.work_format}
            </span>
            {expanded === item.user.id && (
              <UserCard item={item} onAction={(action) => act(item.user.id, action)} />
            )}
          </div>
        ))}

        {items.length < total && (
          <button className="btn btn-ghost" onClick={() => load(items.length)}>
            Показать ещё
          </button>
        )}
      </div>
      <TabBar role="recruiter" />
    </main>
  )
}
