import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { candidateAction, candidates } from '../../api/matching'
import { companyVacancies, myCompanies } from '../../api/company'
import type { CandidateItem, Vacancy } from '../../api/types'
import { WORK_FORMAT_LABELS } from '../../api/types'
import { TabBar } from '../../components/TabBar'
import { Icon } from '../../components/Icon'
import { experienceLabel, shortCity } from '../../components/format'
import { getWebApp } from '../../lib/max'

export default function Feed() {
  const { id } = useParams()
  const vacancyId = Number(id)
  const [items, setItems] = useState<CandidateItem[] | null>(null)
  const [index, setIndex] = useState(0)
  const [vacancy, setVacancy] = useState<Vacancy | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const load = useCallback(() => {
    candidates(vacancyId, 'feed')
      .then((d) => setItems(d.items))
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
      .finally(() => setLoading(false))
  }, [vacancyId])

  useEffect(load, [load])

  useEffect(() => {
    let cancelled = false
    myCompanies()
      .then((cs) => (cs[0] ? companyVacancies(cs[0].id) : []))
      .then((list) => {
        if (!cancelled) setVacancy(list.find((v) => v.id === vacancyId) ?? null)
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [vacancyId])

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

  const current = items?.[index]
  const skills = current
    ? current.resume.skills
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
    : []
  const vacancyChip = vacancy
    ? [vacancy.title, vacancy.city ? shortCity(vacancy.city) : ''].filter(Boolean).join(' · ')
    : ''
  const subParts = current
    ? [
        current.resume.title,
        current.resume.city,
        WORK_FORMAT_LABELS[current.resume.work_format] ?? current.resume.work_format,
        `опыт ${experienceLabel(current.resume.experience_months)}`,
      ].filter(Boolean)
    : []

  return (
    <main className="screen">
      <button className="back" onClick={() => window.history.back()}>
        ← Вакансии
      </button>
      <header className="screen-header" style={{ paddingTop: 8 }}>
        {vacancyChip ? <span className="chip" style={{ padding: '6px 10px', borderRadius: 6, fontSize: 13, fontWeight: 700, color: 'var(--text)' }}>{vacancyChip}</span> : <span />}
        <div className="toggle-group">
          <span className="toggle-chip active">Карточки</span>
          <Link className="toggle-chip" to={`/vacancy/${vacancyId}/list`}>
            Список
          </Link>
        </div>
      </header>
      <div className="screen-body has-tabbar">
        {error && <p className="error-text">{error}</p>}
        {loading && <p className="muted">Загрузка…</p>}

        {!loading && !current && (
          <div className="center-note">
            <Icon name="userBig" size={64} />
            <span>Кандидаты закончились</span>
            <Link className="link-btn" to={`/vacancy/${vacancyId}/list`}>
              Смотреть списком
            </Link>
          </div>
        )}

        {current && (
          <div className="stack-card">
            <div className="photo-area">
              {current.user.photo_url ? (
                <img src={current.user.photo_url} alt="" />
              ) : (
                <Icon name="userBig" size={80} />
              )}
              <span className="match-badge">{Math.round(current.final_score)} Match Score</span>
            </div>
            <div className="stack-body">
              <span className="name">
                {current.user.first_name} {current.user.last_name}
              </span>
              <span className="sub">{subParts.join(' · ')}</span>
              {skills.length > 0 && (
                <div className="stack-tags">
                  {skills.slice(0, 6).map((s) => (
                    <span className="chip" key={s}>
                      {s}
                    </span>
                  ))}
                </div>
              )}
              {current.ai_comment && (
                <div className="ai-block" style={{ marginTop: 8 }}>
                  <span className="ai-label">
                    <Icon name="zap" size={14} />
                    AI резюме
                  </span>
                  <p>«{current.ai_comment}»</p>
                </div>
              )}
            </div>
            <div className="stack-actions">
              <button className="btn btn-ghost btn-lg" onClick={() => act('skip')}>
                Пропустить
              </button>
              <button className="btn btn-lg" onClick={() => act('invite')}>
                Пригласить
              </button>
            </div>
          </div>
        )}
      </div>
      <TabBar role="recruiter" />
    </main>
  )
}
