import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { TabBar } from './TabBar'
import { Icon, type IconName } from './Icon'
import { setRole } from '../api/auth'
import { setStoredUser } from '../lib/session'
import type { MeResponse } from '../api/types'

interface ScreenProps {
  role?: 'candidate' | 'recruiter'
  title?: string
  icon?: IconName
  sub?: string
  children: ReactNode
}

export function Screen({ role, title, icon, sub, children }: ScreenProps) {
  const navigate = useNavigate()

  const switchRole = async () => {
    if (!role) return
    const next = role === 'candidate' ? 'recruiter' : 'candidate'
    try {
      const meData: MeResponse = await setRole(next, true)
      setStoredUser(meData.user)
      navigate('/welcome', { replace: true })
    } catch {
      navigate('/onboarding', { replace: true })
    }
  }

  return (
    <main className="screen">
      {(title || icon || role) && (
        <header className="screen-header">
          {title ? <h1>{title}</h1> : <span />}
          <div className="row" style={{ flexWrap: 'nowrap', gap: 12 }}>
            {icon && (
              <span className="header-icon">
                <Icon name={icon} />
              </span>
            )}
            {role && (
              <button className="header-icon" title="Сменить роль" onClick={switchRole}>
                <Icon name="swap" />
              </button>
            )}
          </div>
        </header>
      )}
      {sub && <p className="screen-sub">{sub}</p>}
      <div className={`screen-body${role ? ' has-tabbar' : ''}`}>{children}</div>
      {role && <TabBar role={role} />}
    </main>
  )
}
