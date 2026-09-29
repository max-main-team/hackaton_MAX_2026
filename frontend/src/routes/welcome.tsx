import { Link } from 'react-router-dom'
import { Screen } from '../components/Screen'
import { Icon } from '../components/Icon'
import { getStoredUser } from '../lib/session'

export default function WelcomeScreen() {
  const role = getStoredUser<{ role: string } | null>()?.role ?? 'candidate'

  const content =
    role === 'recruiter'
      ? {
          icon: 'briefcase' as const,
          title: 'Вы рекрутер',
          text: 'Создайте компанию и вакансии — мы подберём кандидатов и покажем, насколько они вам подходят.',
          cta: 'К компании',
          to: '/company',
        }
      : {
          icon: 'user' as const,
          title: 'Вы кандидат',
          text: 'Компании сами найдут вас. Первым делом заполните резюме — или загрузите PDF, а AI заполнит поля за вас.',
          cta: 'Заполнить резюме',
          to: '/resume',
        }

  return (
    <Screen title="Добро пожаловать!" icon={content.icon}>
      <div className="center-note" style={{ padding: '24px 0' }}>
        <span className="avatar-sq" style={{ width: 56, height: 56 }}>
          <Icon name={content.icon} size={28} />
        </span>
      </div>
      <p style={{ fontSize: 15, lineHeight: 1.5, margin: 0 }}>{content.text}</p>

      <div className="card gap-sm">
        <span className="muted" style={{ fontSize: 12 }}>
          {role === 'candidate'
            ? 'Заполненное резюме появится в подборках компаний — приглашения придут сюда'
            : 'Кандидаты увидят вакансии в ленте подбора — приглашения с дедлайном отправляются в один клик'}
        </span>
      </div>

      <Link className="btn btn-xl" to={content.to}>
        {content.cta}
      </Link>

      <Link className="link-btn" to="/onboarding">
        Сменить роль
      </Link>
    </Screen>
  )
}
