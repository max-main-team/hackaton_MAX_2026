import { useNavigate } from 'react-router-dom'
import { Screen } from '../components/Screen'

export default function ErrorScreen() {
  const navigate = useNavigate()

  return (
    <Screen title="Что-то сломалось">
      <div className="status-card down">
        <span className="status-title">Ошибка</span>
        <span className="status-note">Мы уже в курсе. Попробуйте перезагрузить мини-приложение.</span>
      </div>
      <button className="btn" onClick={() => window.location.reload()}>
        Перезагрузить
      </button>
      <button className="btn btn-ghost" onClick={() => navigate('/')}>
        На главный экран
      </button>
    </Screen>
  )
}
