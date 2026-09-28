import { useEffect } from 'react'
import { Navigate } from 'react-router-dom'
import type { ReactNode } from 'react'
import { getToken, getStoredUser } from '../lib/session'
import { callBridge } from '../lib/max'

export function homeFor(role: string | null | undefined): string {
  if (role === 'candidate') return '/resume'
  if (role === 'recruiter') return '/company'
  return '/onboarding'
}

interface GuardProps {
  role?: 'candidate' | 'recruiter'
  /** разрешить доступ без выбранной роли (нужно самому онбордингу) */
  allowNoRole?: boolean
  children: ReactNode
}

export function Guard({ role, allowNoRole, children }: GuardProps) {
  useEffect(() => {
    callBridge((app) => app.enableClosingConfirmation?.())
  }, [])

  if (!getToken()) {
    return <Navigate to="/" replace />
  }

  const user = getStoredUser<{ role: string } | null>()
  if ((!user || !user.role) && !allowNoRole) {
    return <Navigate to="/onboarding" replace />
  }
  if (role && user.role !== role) {
    return <Navigate to={homeFor(user.role)} replace />
  }
  return <>{children}</>
}
