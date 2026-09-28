import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { setRole } from '../api/auth'
import { setStoredUser } from '../lib/session'
import { Icon } from '../components/Icon'
import type { MeResponse } from '../api/types'

export default function Onboarding() {
  const navigate = useNavigate()
  const [consent, setConsent] = useState(false)
  const [busy, setBusy] = useState<string | null>(null)
  const [error, setError] = useState('')

  const choose = async (role: 'candidate' | 'recruiter') => {
    if (!consent) {
      setError('Подтвердите согласие на обработку персональных данных')
      return
    }
    setBusy(role)
    setError('')
    try {
      const meData: MeResponse = await setRole(role, true)
      setStoredUser(meData.user)
      navigate(role === 'candidate' ? '/resume' : '/company', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(null)
    }
  }

  return (
    <main className="screen">
      <header className="screen-header">
        <h1>Кто вы?</h1>
      </header>
      <p className="screen-sub">Это определит ваш сценарий работы</p>

      <div className="screen-body" style={{ paddingTop: 16, gap: 14 }}>
        <button
          className="card"
          style={{ cursor: 'pointer', textAlign: 'left', alignItems: 'flex-start' }}
          onClick={() => choose('candidate')}
          disabled={busy !== null}
        >
          <div className="row" style={{ flexWrap: 'nowrap', gap: 12 }}>
            <span className="avatar-sq" style={{ width: 40, height: 40 }}>
              <Icon name="user" size={20} />
            </span>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <strong style={{ fontSize: 15 }}>Я кандидат</strong>
              <span className="muted" style={{ fontSize: 12 }}>
                Ищу работу — компании сами найдут меня
              </span>
            </div>
          </div>
        </button>

        <button
          className="card"
          style={{ cursor: 'pointer', textAlign: 'left', alignItems: 'flex-start' }}
          onClick={() => choose('recruiter')}
          disabled={busy !== null}
        >
          <div className="row" style={{ flexWrap: 'nowrap', gap: 12 }}>
            <span className="avatar-sq" style={{ width: 40, height: 40 }}>
              <Icon name="briefcase" size={20} />
            </span>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <strong style={{ fontSize: 15 }}>Я рекрутер</strong>
              <span className="muted" style={{ fontSize: 12 }}>
                Ищу кандидатов — система подберёт приоритеты
              </span>
            </div>
          </div>
        </button>

        <label className="row" style={{ gap: 8, fontSize: 13 }}>
          <input
            type="checkbox"
            checked={consent}
            onChange={(e) => setConsent(e.target.checked)}
            style={{ accentColor: 'var(--accent)' }}
          />
          <span className="muted">Согласен на обработку персональных данных (резюме, профиль MAX)</span>
        </label>

        {error && <p className="error-text">{error}</p>}
      </div>
    </main>
  )
}
