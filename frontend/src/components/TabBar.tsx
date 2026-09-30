import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { companyVacancies, myCompanies } from '../api/company'
import type { Company } from '../api/types'
import { Icon } from './Icon'

interface Tab {
  to: string
  label: string
  icon: Parameters<typeof Icon>[0]['name']
  match: (path: string) => boolean
}

const CANDIDATE_TABS: Tab[] = [
  { to: '/resume', label: 'Резюме', icon: 'user', match: (p) => p === '/resume' },
  { to: '/invitations', label: 'Приглашения', icon: 'mail', match: (p) => p === '/invitations' },
  { to: '/map', label: 'Карта', icon: 'map', match: (p) => p === '/map' },
  { to: '/referral', label: 'Пригласить', icon: 'plus', match: (p) => p === '/referral' },
]

const RECRUITER_TABS: Tab[] = [
  { to: '/company', label: 'Компания', icon: 'briefcase', match: (p) => p === '/company' },
  {
    to: '/company/vacancies',
    label: 'Вакансии',
    icon: 'list',
    match: (p) => p === '/company/vacancies' || p === '/company/vacancies/new' || p === '/vacancy-list',
  },
  {
    to: '/company/candidates',
    label: 'Кандидаты',
    icon: 'userBig',
    match: (p) => p.startsWith('/company/candidates'),
  },
  { to: '/match', label: 'Подбор', icon: 'zap', match: (p) => p.startsWith('/vacancy/') },
  { to: '/map', label: 'Карта', icon: 'map', match: (p) => p === '/map' },
]

export function TabBar({ role }: { role: 'candidate' | 'recruiter' }) {
  const location = useLocation()
  const [feedPath, setFeedPath] = useState<string | null>(null)

  useEffect(() => {
    if (role !== 'recruiter') return
    let cancelled = false
    myCompanies()
      .then((cs: Company[]) => {
        const company = cs[0]
        if (!company) return
        return companyVacancies(company.id).then((vacancies) => {
          if (cancelled) return
          const vacancy = vacancies.find((v) => v.is_active) ?? vacancies[0] ?? null
          setFeedPath(vacancy ? `/vacancy/${vacancy.id}/feed` : '/company/vacancies')
        })
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [role, location.pathname])

  const tabs = (role === 'candidate' ? CANDIDATE_TABS : RECRUITER_TABS).map((tab) =>
    tab.label === 'Подбор' && feedPath ? { ...tab, to: feedPath } : tab,
  )

  return (
    <nav className="tabbar">
      {tabs.map((tab) => (
        <Link
          key={tab.label}
          className={`tabbar-item${tab.match(location.pathname) ? ' active' : ''}`}
          to={tab.to}
        >
          <Icon name={tab.icon} />
          <span>{tab.label}</span>
        </Link>
      ))}
    </nav>
  )
}
