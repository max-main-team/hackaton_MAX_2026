import type { ReactNode } from 'react'
import { TabBar } from './TabBar'
import { Icon, type IconName } from './Icon'

interface ScreenProps {
  role?: 'candidate' | 'recruiter'
  title?: string
  icon?: IconName
  sub?: string
  children: ReactNode
}

export function Screen({ role, title, icon, sub, children }: ScreenProps) {
  return (
    <main className="screen">
      {(title || icon) && (
        <header className="screen-header">
          {title ? <h1>{title}</h1> : <span />}
          {icon && (
            <span className="header-icon">
              <Icon name={icon} />
            </span>
          )}
        </header>
      )}
      {sub && <p className="screen-sub">{sub}</p>}
      <div className={`screen-body${role ? ' has-tabbar' : ''}`}>{children}</div>
      {role && <TabBar role={role} />}
    </main>
  )
}
