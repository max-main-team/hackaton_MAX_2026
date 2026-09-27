import { useEffect } from 'react'
import { Navigate } from 'react-router-dom'
import type { ReactNode } from 'react'
import { getToken, getStoredUser } from '../lib/session'
import { getWebApp } from '../lib/max'

export function homeFor(role: string | null | undefined): string {
  if (role === 'candidate') return '/resume'
  if (role === 'recruiter') return '/company'
  return '/onboarding'
}

interface GuardProps {
  role?: 'candidate' | 'recruiter'
  children: ReactNode
}

export function Guard({ role, children }: GuardProps) {
  useEffect(() => {
    const app = getWebApp()
    app?.ready()
    app?.expand()
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
  void location
  return <>{children}</>
}
