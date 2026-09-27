import { useEffect, useState } from 'react'
import './App.css'
import { api } from './lib/api'

function App() {
  const [health, setHealth] = useState<'checking' | 'ok' | 'down'>('checking')

  useEffect(() => {
    api.health()
      .then(() => setHealth('ok'))
      .catch(() => setHealth('down'))
  }, [])

  return (
    <main className="app">
      <h1>MAX Mini App</h1>
      <p className="subtitle">Реверс-найм внутри MAX · хакатон 2026</p>

      {health === 'checking' && <p className="status-line">Проверяем сервис…</p>}

      {health === 'ok' && (
        <section className="status-card ok">
          <span className="status-title">✅ Всё работает</span>
          <span className="status-note">API и база данных доступны</span>
        </section>
      )}

      {health === 'down' && (
        <section className="status-card down">
          <span className="status-title">❌ Сервис недоступен</span>
          <span className="status-note">Уже чиним, загляни чуть позже</span>
        </section>
      )}

      <footer className="footer">eclipse-sim.ru · CI/CD: GitHub Actions</footer>
    </main>
  )
}

export default App
