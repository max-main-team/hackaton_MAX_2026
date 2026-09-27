import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { candidateAction, candidates } from '../../api/matching'
import type { CandidateItem } from '../../api/types'
import { getWebApp } from '../../lib/max'

export default function Feed() {
  const { id } = useParams()
  const vacancyId = Number(id)
  const [items, setItems] = useState<CandidateItem[] | null>(null)
  const [index, setIndex] = useState(0)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    candidates(vacancyId, 'feed')
      .then((d) => setItems(d.items))
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
      .finally(() => setLoading(false))
  }, [vacancyId])

  useEffect(load, [load])

  const act = async (action: 'invite' | 'skip') => {
    const current = items?.[index]
    if (!current) return
    getWebApp()?.HapticFeedback?.impactOccurred('light')
    try {
      await candidateAction(vacancyId, current.user.id, action)
      setIndex((i) => i + 1)
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

  const current = items?.[index]

  return (
    <main className="page">
      <Link className="back" to="/company/vacancies">
        ← Вакансии
      </Link>
      <h1 className="page-title">Лента кандидатов</h1>
      <p className="page-sub">Свайп вправо — пригласить, влево — пропустить</p>

      {error && <p className="error-text">{error}</p>}

      {current ? (
        <div className="card" style={{ alignItems: 'center', textAlign: 'center' }}>
          <div style={{ fontSize: 44 }}>{current.user.photo_url ? '👤' : '👤'}</div>
          <strong style={{ fontSize: 20 }}>
            {current.user.first_name} {current.user.last_name}
          </strong>
          <span className="tag">score {current.score}</span>
          <span className="muted">{current.resume.title}</span>
          <div className="row" style={{ justifyContent: 'center' }}>
            {current.resume.skills
              .split(',')
              .map((s) => s.trim())
              .filter(Boolean)
              .slice(0, 6)
              .map((s) => (
                <span className="tag" key={s}>
                  {s}
                </span>
              ))}
          </div>
          <span className="muted">
            {current.resume.city} · опыт {current.resume.experience_months} мес
          </span>
          {current.ai_comment && <span className="muted">AI: {current.ai_comment}</span>}
          <div className="row" style={{ justifyContent: 'center', marginTop: 8 }}>
            <button className="btn btn-danger" style={{ fontSize: 20 }} onClick={() => act('skip')}>
              ✕
            </button>
            <button className="btn btn-success" style={{ fontSize: 20 }} onClick={() => act('invite')}>
              ♥
            </button>
          </div>
        </div>
      ) : (
        <p className="muted">Кандидаты закончились</p>
      )}

      <div className="row" style={{ justifyContent: 'center' }}>
        <Link className="link-btn" to={`/vacancy/${vacancyId}/list`}>
          Смотреть списком
        </Link>
      </div>
    </main>
  )
}
