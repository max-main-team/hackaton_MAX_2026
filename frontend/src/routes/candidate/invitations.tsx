import { useCallback, useEffect, useState } from 'react'
import { respondInvitation, invitations } from '../../api/matching'
import type { Invitation, MatchResult } from '../../api/types'
import { WORK_FORMAT_LABELS } from '../../api/types'
import { VerifiedBadge } from '../../components/Badge'
import { Countdown } from '../../components/Countdown'
import { Screen } from '../../components/Screen'
import { Icon } from '../../components/Icon'
import { formatSalary } from '../../components/format'
import { getWebApp } from '../../lib/max'

function companyInitials(name: string): string {
  const words = name.trim().split(/\s+/).slice(0, 2)
  return words.map((w) => w.charAt(0).toUpperCase()).join('')
}

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
      <Screen role="candidate" title="Матч!" icon="bell">
        <div className="status-card ok" style={{ alignItems: 'center', textAlign: 'center' }}>
          <span className="status-title">🎉 {match.company.name}</span>
          <span>{match.vacancy.title}</span>
          <span className="status-note">
            Свяжитесь с рекрутером: @{match.recruiter_contact || 'контакт уточняется'}
          </span>
        </div>
        <button className="btn btn-ghost" onClick={load}>
          К другим приглашениям
        </button>
      </Screen>
    )
  }

  return (
    <Screen role="candidate" title="Приглашения" icon="bell">
      {error && <p className="error-text">{error}</p>}
      {list === null && <p className="muted">Загрузка…</p>}
      {list !== null && list.length === 0 && (
        <div className="center-note">
          <Icon name="mail" size={48} />
          <span>Пока нет приглашений</span>
        </div>
      )}

      {(list ?? []).map((inv) => (
        <div className="card gap-sm" key={inv.id}>
          <div className="row between" style={{ flexWrap: 'nowrap' }}>
            <div className="row" style={{ flexWrap: 'nowrap', minWidth: 0 }}>
              <span className="avatar-sq">{companyInitials(inv.company.name)}</span>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 2, minWidth: 0 }}>
                <div className="row" style={{ flexWrap: 'nowrap', gap: 4 }}>
                  <span style={{ fontSize: 14, fontWeight: 700, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {inv.company.name}
                  </span>
                  <VerifiedBadge verified={inv.company.verified} />
                </div>
                <span className="muted" style={{ fontSize: 12 }}>
                  {inv.vacancy.city || ''}
                  {inv.vacancy.city && inv.vacancy.work_format ? ' · ' : ''}
                  {WORK_FORMAT_LABELS[inv.vacancy.work_format] ?? ''}
                </span>
              </div>
            </div>
            {inv.status === 'pending' && <Countdown deadlineAt={inv.deadline_at} hoursLeft={inv.hours_left} />}
            {inv.status === 'overdue' && <span className="timer overdue">Просрочено</span>}
            {inv.status === 'responded' && (
              <span className="chip">{inv.response === 'accept' ? 'Принято' : 'Отклонено'}</span>
            )}
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <span className="vacancy-role">{inv.vacancy.title}</span>
            {formatSalary(inv.vacancy.salary_min, inv.vacancy.salary_max) && (
              <span className="salary">{formatSalary(inv.vacancy.salary_min, inv.vacancy.salary_max)}</span>
            )}
          </div>
          {inv.status === 'pending' && (
            <div className="row" style={{ flexWrap: 'nowrap' }}>
              <button className="btn" style={{ flex: 1 }} onClick={() => respond(inv.id, 'accept')}>
                Принять
              </button>
              <button className="btn btn-ghost" style={{ flex: 1 }} onClick={() => respond(inv.id, 'decline')}>
                Отклонить
              </button>
            </div>
          )}
        </div>
      ))}
    </Screen>
  )
}
