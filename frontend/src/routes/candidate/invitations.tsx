import { useCallback, useEffect, useState } from 'react'
import { respondInvitation, invitations } from '../../api/matching'
import type { Invitation, MatchResult } from '../../api/types'
import { Badge } from '../../components/Badge'
import { Countdown } from '../../components/Countdown'
import { getWebApp } from '../../lib/max'

export default function InvitationsScreen() {
  const [list, setList] = useState<Invitation[] | null>(null)
  const [error, setError] = useState('')
  const [match, setMatch] = useState<MatchResult | null>(null)

  const load = useCallback(() => {
    invitations()
      .then(setList)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  useEffect(load, [load])

  const respond = async (id: number, response: 'accept' | 'decline') => {
    setError('')
    try {
      const result = await respondInvitation(id, response)
      if (response === 'accept') {
        getWebApp()?.HapticFeedback?.notificationOccurred('success')
        setMatch(result)
      }
      load()
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }

  if (match) {
    return (
      <main className="page">
        <div className="status-card ok" style={{ alignItems: 'center', textAlign: 'center' }}>
          <span className="status-title">🎉 Матч!</span>
          <span>
            {match.company.name} · {match.vacancy.title}
          </span>
          <span className="muted">
            Свяжитесь с рекрутером: @{match.recruiter_contact || 'контакт уточняется'}
          </span>
        </div>
        <button className="btn btn-secondary" onClick={load}>
          К другим приглашениям
        </button>
      </main>
    )
  }

  return (
    <main className="page">
      <h1 className="page-title">Приглашения</h1>

      {error && <p className="error-text">{error}</p>}
      {list === null && <p className="muted">Загрузка…</p>}
      {list !== null && list.length === 0 && <p className="muted">Пока нет приглашений</p>}

      {(list ?? []).map((inv) => (
        <div className="list-item" key={inv.id}>
          <div className="row" style={{ justifyContent: 'space-between' }}>
            <strong>{inv.company.name}</strong>
            <Badge verified={inv.company.verified} />
          </div>
          <span>
            {inv.vacancy.title}
            {inv.vacancy.city ? ` · ${inv.vacancy.city}` : ''}
          </span>
          {inv.status === 'pending' && <Countdown deadlineAt={inv.deadline_at} />}
          {inv.status === 'overdue' && <span className="error-text">Просрочено</span>}
          {inv.status === 'responded' && <span className="muted">Ответ отправлен: {inv.response}</span>}

          {inv.status === 'pending' && (
            <div className="row">
              <button className="btn btn-success" onClick={() => respond(inv.id, 'accept')}>
                Принять
              </button>
              <button className="btn btn-secondary" onClick={() => respond(inv.id, 'decline')}>
                Отклонить
              </button>
            </div>
          )}
        </div>
      ))}
    </main>
  )
}
