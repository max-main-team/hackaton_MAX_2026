import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { setRole } from '../api/auth'
import { setStoredUser } from '../lib/session'
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
    <main className="page">
      <h1 className="page-title">Кто вы?</h1>
      <p className="page-sub">Это определит ваш сценарий работы</p>

      <button
        className="card"
        style={{ cursor: 'pointer', textAlign: 'left', border: 'none' }}
        onClick={() => choose('candidate')}
        disabled={busy !== null}
      >
        <strong>Я кандидат</strong>
        <span className="muted">Ищу работу — компании сами найдут меня</span>
      </button>

      <button
        className="card"
        style={{ cursor: 'pointer', textAlign: 'left', border: 'none' }}
        onClick={() => choose('recruiter')}
        disabled={busy !== null}
      >
        <strong>Я рекрутер</strong>
        <span className="muted">Ищу кандидатов — система подберёт приоритеты</span>
      </button>

      <label className="row" style={{ gap: 8, fontSize: 13 }}>
        <input type="checkbox" checked={consent} onChange={(e) => setConsent(e.target.checked)} />
        <span className="muted">
          Согласен на обработку персональных данных (резюме, профиль MAX)
        </span>
      </label>

      {error && <p className="error-text">{error}</p>}
    </main>
  )
}
