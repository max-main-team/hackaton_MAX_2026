import { useCallback, useEffect, useState } from 'react'
import { myReferrals } from '../../api/auth'
import type { ReferralsResponse } from '../../api/types'
import { getWebApp } from '../../lib/max'
import { getStoredUser } from '../../lib/session'

const BOT_LINK = 'https://max.ru/t599_hakaton_max_bot?startapp='

export default function ReferralScreen() {
  const [data, setData] = useState<ReferralsResponse | null>(null)
  const [code, setCode] = useState('')
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    myReferrals()
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  useEffect(() => {
    const user = getStoredUser<{ id: number } | null>()
    if (user) setCode(`ref_${user.id}`)
    load()
  }, [load])

  const link = BOT_LINK + code

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(link)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      setError('Не удалось скопировать')
    }
  }

  const share = () => {
    const app = getWebApp()
    if (app?.shareMaxContent) {
      app.shareMaxContent({ text: `Ищи работу по-новому: ${link}` })
    } else if (navigator.share) {
      navigator.share({ text: `Ищи работу по-новому: ${link}` }).catch(() => {})
    } else {
      copy()
    }
  }

  return (
    <main className="page">
      <h1 className="page-title">Пригласить друзей</h1>
      <p className="page-sub">За каждого друга, который придёт по вашей ссылке</p>

      {error && <p className="error-text">{error}</p>}
      {data && (
        <div className="status-card ok">
          <span className="status-title">Приглашено: {data.count}</span>
          {data.items.length > 0 && (
            <span className="status-note">
              {data.items.map((i) => i.first_name).join(', ')}
            </span>
          )}
        </div>
      )}

      <div className="card">
        <span className="muted">Ваша ссылка</span>
        <span style={{ wordBreak: 'break-all', fontSize: 13 }}>{link}</span>
        <div className="row">
          <button className="btn" onClick={copy}>
            {copied ? 'Скопировано ✓' : 'Скопировать'}
          </button>
          <button className="btn btn-secondary" onClick={share}>
            Поделиться
          </button>
        </div>
      </div>
    </main>
  )
}
