import { useCallback, useEffect, useState } from 'react'
import './App.css'
import { getInitData, getWebApp, isInsideMax } from './lib/max'
import { api, type AuthUser } from './lib/api'

function App() {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    getWebApp()?.ready()
    getWebApp()?.expand()
  }, [])

  const login = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      setUser(await api.auth(getInitData()))
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }, [])

  return (
    <main className="app">
      <h1>MAX Mini App</h1>

      {!isInsideMax() && (
        <p className="hint">
          Открыто вне MAX — initData недоступна, авторизация вернёт ошибку.
        </p>
      )}

      {user ? (
        <section className="card">
          <h2>
            {user.first_name} {user.last_name}
          </h2>
          <dl>
            <dt>ID</dt>
            <dd>{user.id}</dd>
            <dt>Username</dt>
            <dd>{user.username || '—'}</dd>
          </dl>
        </section>
      ) : (
        <button onClick={login} disabled={loading}>
          {loading ? 'Входим…' : 'Войти через MAX'}
        </button>
      )}

      {error && <p className="error">{error}</p>}
    </main>
  )
}

export default App
