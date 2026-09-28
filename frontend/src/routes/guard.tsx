import { useEffect } from 'react'
import { Navigate } from 'react-router-dom'
import type { ReactNode } from 'react'
import { getToken, getStoredUser } from '../lib/session'
import { callBridge } from '../lib/max'
import { homeFor } from './home'

export function Guard({ role, children }: GuardProps) {
  useEffect(() => {
    callBridge((app) => app.enableClosingConfirmation?.())
  }, [])

  if (!getToken()) {
    return <Navigate to="/" replace />
  }

  const user = getStoredUser<{ role: string } | null>()
  if (!user || !user.role) {
    return <Navigate to="/onboarding" replace />
  }
  if (role && user.role !== role) {
    return <Navigate to={homeFor(user.role)} replace />
  }
  return <>{children}</>
}

interface GuardProps {
  role?: 'candidate' | 'recruiter'
  children: ReactNode
}

