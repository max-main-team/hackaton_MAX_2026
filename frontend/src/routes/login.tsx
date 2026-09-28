import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { auth, demoAuth, me } from '../api/auth'
import { getInitData } from '../lib/max'
import { getToken, setStoredUser, setToken } from '../lib/session'
import type { MeResponse } from '../api/types'

function homeRedirect(role: string): string {
  return role === 'candidate' ? '/resume' : role === 'recruiter' ? '/company' : '/onboarding'
}

export default function Login() {
  const navigate = useNavigate()
  const [status, setStatus] = useState<'idle' | 'checking' | 'error'>('idle')
  const [error, setError] = useState('')
  const tried = useRef(false)

  const applyMe = async () => {
    const meData: MeResponse = await me()
    setStoredUser(meData.user)
    navigate(homeRedirect(meData.user.role), { replace: true })
  }

  useEffect(() => {
    if (tried.current) return
    tried.current = true

    if (getToken()) {
      applyMe().catch(() => {
        setStatus('error')
        setError('Сессия истекла, войдите заново')
      })
      return
    }

    const initData = getInitData()
    if (!initData) {
      setStatus('error')
      setError('Откройте мини-приложение через MAX')
      return
    }
    setStatus('checking')
    auth(initData)
      .then((res) => {
        setToken(res.token)
        setStoredUser(res.user)
        return applyMe()
      })
      .catch((e) => {
        setStatus('error')
        setError(e instanceof Error ? e.message : String(e))
      })
  }, [])

  const demoLogin = () => {
    setStatus('checking')
    demoAuth()
      .then((res) => {
        setToken(res.token)
        setStoredUser(res.user)
        navigate(homeRedirect(res.user.role), { replace: true })
      })
      .catch((e) => {
        setStatus('error')
        setError(e instanceof Error ? e.message : 'Ошибка входа')
      })
  }

  return (
    <main className="page">
      <h1 className="page-title">MAX Mini App</h1>
      <p className="page-sub">Реверс-найм внутри MAX · хакатон 2026</p>

      {status === 'checking' && <p className="muted">Входим…</p>}

      {status === 'error' && (
        <>
          <div className="status-card down">
            <span className="status-title">Не удалось войти</span>
            <span className="status-note">{error}</span>
          </div>
          <button className="btn" onClick={demoLogin}>
            Демо-вход
          </button>
        </>
      )}

      {status === 'idle' && (
        <button className="btn" onClick={demoLogin}>
          Войти
        </button>
      )}

      <footer className="muted" style={{ marginTop: 'auto', textAlign: 'center' }}>
        eclipse-sim.ru · CI/CD: GitHub Actions
      </footer>
    </main>
  )
}
