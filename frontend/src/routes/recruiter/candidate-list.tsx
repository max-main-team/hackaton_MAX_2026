import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { candidateAction, candidates } from '../../api/matching'
import type { CandidateItem } from '../../api/types'
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
    <main className="page">
      <Link className="back" to="/company/vacancies">
        ← Вакансии
      </Link>
      <h1 className="page-title">Кандидаты ({total})</h1>

      {error && <p className="error-text">{error}</p>}
      {loading && <p className="muted">Загрузка…</p>}

      {!loading && items.length === 0 && <p className="muted">Все кандидаты разобраны</p>}

      {items.map((item) => (
        <div className="list-item" key={item.user.id}>
          <div className="row" style={{ justifyContent: 'space-between', cursor: 'pointer' }} onClick={() => setExpanded(expanded === item.user.id ? null : item.user.id)}>
            <strong>
              {item.user.first_name} {item.user.last_name}
            </strong>
            <span className="badge">{item.score}</span>
          </div>
          <span className="muted">
            {item.resume.title} · {item.resume.city} · {WORK(item.resume.work_format)}
          </span>
          {expanded === item.user.id && (
            <UserCard
              item={item}
              onAction={(action) => act(item.user.id, action)}
            />
          )}
        </div>
      ))}

      {items.length < total && (
        <button className="btn btn-secondary" onClick={() => load(items.length)}>
          Показать ещё
        </button>
      )}
    </main>
  )
}

function WORK(format: string): string {
  return { onsite: 'офис', hybrid: 'гибрид', remote: 'удалённо' }[format] ?? format
}
