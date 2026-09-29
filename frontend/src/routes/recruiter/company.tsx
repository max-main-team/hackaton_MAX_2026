import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { companyReferrals, createCompany, myCompanies, verifyCompany } from '../../api/company'
import type { Company, CompanyReferralsResponse } from '../../api/types'
import { Screen } from '../../components/Screen'
import { VerifiedBadge } from '../../components/Badge'
import { Icon } from '../../components/Icon'
import { getWebApp } from '../../lib/max'

const BOT_LINK = 'https://max.ru/t599_hakaton_max_bot?startapp='

export default function CompanyScreen() {
  const [companies, setCompanies] = useState<Company[] | null>(null)
  const [form, setForm] = useState({
    name: '',
    description: '',
    website: '',
    logo_url: '',
    address: '',
    position: 'owner',
  })
  const [error, setError] = useState('')
  const [referrals, setReferrals] = useState<CompanyReferralsResponse | null>(null)
  const [copied, setCopied] = useState(false)
  const [botToken, setBotToken] = useState('')
  const [verifying, setVerifying] = useState(false)

  const load = useCallback(() => {
    myCompanies()
      .then(setCompanies)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  useEffect(load, [load])

  const verify = async (companyID: number) => {
    setError('')
    setVerifying(true)
    try {
      await verifyCompany(companyID, botToken.trim())
      setBotToken('')
      load()
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setVerifying(false)
    }
  }

  useEffect(() => {
    if (companies?.[0]) {
      companyReferrals()
        .then(setReferrals)
        .catch(() => {})
    }
  }, [companies])

  const referralLink = companies?.[0] ? BOT_LINK + `refc_${companies[0].id}` : ''

  const copyLink = async () => {
    try {
      await navigator.clipboard.writeText(referralLink)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      setError('Не удалось скопировать')
    }
  }

  const shareLink = () => {
    const app = getWebApp()
    if (app?.shareMaxContent) {
      app.shareMaxContent({ text: `Ищем HR в команду платформы: ${referralLink}` })
    } else if (navigator.share) {
      navigator.share({ text: `Ищем HR в команду платформы: ${referralLink}` }).catch(() => {})
    } else {
      copyLink()
    }
  }

  const create = async () => {
    setError('')
    try {
      await createCompany(form)
      load()
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }

  if (companies === null) {
    return (
      <Screen role="recruiter">
        <p className="muted">Загрузка…</p>
      </Screen>
    )
  }

  const current = companies[0]

  return (
    <Screen role="recruiter" title="Моя компания" icon="briefcase">
      {error && <p className="error-text">{error}</p>}

      {current ? (
        <>
          <div className="card gap-sm">
            <div className="row between" style={{ flexWrap: 'nowrap' }}>
              <span className="avatar-sq">{current.name.charAt(0).toUpperCase()}</span>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: 1, minWidth: 0 }}>
                <div className="row" style={{ flexWrap: 'nowrap', gap: 4 }}>
                  <strong style={{ fontSize: 15, overflow: 'hidden', textOverflow: 'ellipsis' }}>{current.name}</strong>
                  <VerifiedBadge verified={current.verified} />
                </div>
                {current.website && (
                  <a className="muted" style={{ fontSize: 12 }} href={current.website} target="_blank" rel="noreferrer">
                    {current.website}
                  </a>
                )}
              </div>
            </div>
            {current.description && <span className="muted">{current.description}</span>}
            {current.address && <span className="muted">{current.address}</span>}
          </div>
          <Link className="btn" to="/company/vacancies">
            Мои вакансии
          </Link>

          {!current.verified && (
            <div className="form-section">
              <p className="section-label">Верификация в MAX</p>
              <div className="card gap-sm">
                <span className="muted" style={{ fontSize: 12 }}>
                  Вставьте токен бота вашей компании — получите бейдж доверия для кандидатов
                </span>
                <div className="field">
                  <label>Токен бота (business.max.ru → Чат-боты → Настройки)</label>
                  <input
                    className="input"
                    type="password"
                    placeholder="Токен бота"
                    value={botToken}
                    onChange={(e) => setBotToken(e.target.value)}
                  />
                </div>
                <button
                  className="btn"
                  onClick={() => verify(current.id)}
                  disabled={verifying || !botToken.trim()}
                >
                  {verifying ? 'Проверяем…' : 'Верифицировать'}
                </button>
              </div>
            </div>
          )}

          <div className="form-section">
            <p className="section-label">Пригласить HR</p>
            <div className="card gap-sm">
              {referrals && (
                <div className="row between">
                  <span className="muted" style={{ fontSize: 12 }}>Осталось приглашений</span>
                  <span className="chip accent">
                    {Math.max(0, referrals.invite_quota - referrals.invite_used)} из {referrals.invite_quota}
                  </span>
                </div>
              )}
              {referrals?.promo_until && new Date(referrals.promo_until) > new Date() && (
                <span className="timer">🏷 Промо до {new Date(referrals.promo_until).toLocaleDateString('ru-RU')}</span>
              )}
              {referrals && referrals.invited.length > 0 && (
                <div className="stack-tags">
                  {referrals.invited.map((ic) => (
                    <span className="chip" key={ic.id}>
                      {ic.name}
                    </span>
                  ))}
                </div>
              )}
              <span className="muted" style={{ fontSize: 12 }}>
                Пригласите HR из другой компании: они получают промо на 30 дней, вы — +5 приглашений
              </span>
              <span style={{ wordBreak: 'break-all', fontSize: 13 }}>{referralLink}</span>
              <div className="row">
                <button className="btn" style={{ flex: 1 }} onClick={copyLink}>
                  {copied ? 'Скопировано ✓' : 'Скопировать'}
                </button>
                <button className="btn btn-ghost" style={{ flex: 1 }} onClick={shareLink}>
                  <Icon name="plus" size={14} />
                  Поделиться
                </button>
              </div>
            </div>
          </div>
        </>
      ) : (
        <>
          <div className="form-section">
            <p className="section-label">Новая компания</p>
            <div className="fieldset">
              <div className="field">
                <label>Название *</label>
                <input className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
              </div>
              <div className="field">
                <label>Описание</label>
                <textarea
                  className="textarea"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                />
              </div>
              <div className="field">
                <label>Сайт</label>
                <input
                  className="input"
                  value={form.website}
                  onChange={(e) => setForm({ ...form, website: e.target.value })}
                />
              </div>
              <div className="field">
                <label>Адрес</label>
                <input
                  className="input"
                  value={form.address}
                  onChange={(e) => setForm({ ...form, address: e.target.value })}
                />
              </div>
              <div className="field">
                <label>Ваша позиция</label>
                <select
                  className="select"
                  value={form.position}
                  onChange={(e) => setForm({ ...form, position: e.target.value })}
                >
                  <option value="owner">Директор / владелец</option>
                  <option value="hr">HR</option>
                  <option value="employee">Сотрудник</option>
                </select>
              </div>
            </div>
          </div>

          <button className="btn btn-xl" onClick={create} disabled={!form.name.trim()}>
            Создать компанию
          </button>

          <p className="muted" style={{ fontSize: 12, margin: 0 }}>
            Верификация компании появится здесь же — понадобится токен бота компании.
          </p>
        </>
      )}
    </Screen>
  )
}
