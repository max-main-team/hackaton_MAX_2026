import { useNavigate } from 'react-router-dom'
import { callBridge } from '../lib/max'

export default function ErrorScreen() {
  const navigate = useNavigate()

  return (
    <main className="page">
      <div className="status-card down">
        <span className="status-title">Что-то сломалось</span>
        <span className="status-note">Мы уже в курсе. Попробуйте перезагрузить мини-приложение.</span>
      </div>
      <button className="btn" onClick={() => window.location.reload()}>
        Перезагрузить
      </button>
      <button
        className="btn btn-secondary"
        onClick={() => {
          callBridge((app) => app.close?.())
          navigate('/')
        }}
      >
        На главный экран
      </button>
    </main>
  )
}
