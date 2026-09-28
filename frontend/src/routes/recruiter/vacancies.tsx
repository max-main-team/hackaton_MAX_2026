import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { companyVacancies, myCompanies, updateVacancy } from '../../api/company'
import type { Company, Vacancy } from '../../api/types'
import { WORK_FORMAT_LABELS } from '../../api/types'
import { Screen } from '../../components/Screen'
import { formatSalary } from '../../components/format'

export default function VacanciesScreen() {
  const [company, setCompany] = useState<Company | null>(null)
  const [vacancies, setVacancies] = useState<Vacancy[] | null>(null)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    myCompanies()
      .then((cs) => {
        setCompany(cs[0] ?? null)
        if (cs[0]) return companyVacancies(cs[0].id).then(setVacancies)
        setVacancies([])
      })
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  useEffect(load, [])

  const toggle = async (v: Vacancy) => {
    try {
      const updated = await updateVacancy(v.id, { is_active: !v.is_active })
      setVacancies((list) => (list ?? []).map((x) => (x.id === updated.id ? updated : x)))
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }

  if (vacancies === null) {
    return (
      <Screen role="recruiter">
        <p className="muted">Загрузка…</p>
      </Screen>
    )
  }

  return (
    <Screen role="recruiter" title="Вакансии" icon="list" sub={company ? company.name : undefined}>
      {error && <p className="error-text">{error}</p>}

      <Link className="btn" to="/company/vacancies/new">
        + Новая вакансия
      </Link>

      {vacancies.length === 0 && (
        <div className="center-note">
          <span>Пока нет вакансий</span>
        </div>
      )}

      {vacancies.map((v) => (
        <div className="card gap-sm" key={v.id}>
          <div className="row between" style={{ flexWrap: 'nowrap' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 2, minWidth: 0 }}>
              <strong style={{ fontSize: 15, overflow: 'hidden', textOverflow: 'ellipsis' }}>{v.title}</strong>
              <span className="muted" style={{ fontSize: 12 }}>
                {WORK_FORMAT_LABELS[v.work_format] ?? v.work_format} · {v.city || '—'} · TTL {v.response_ttl_hours}ч
              </span>
            </div>
            <span className={`chip${v.is_active ? '' : ' '}`} style={v.is_active ? { color: 'var(--green)', borderColor: 'var(--ok-border)' } : { color: 'var(--warn-text)', background: 'var(--warn-bg)', borderColor: 'var(--warn-border)' }}>
              {v.is_active ? 'Активна' : 'Архив'}
            </span>
          </div>
          {formatSalary(v.salary_min, v.salary_max) && <span className="salary">{formatSalary(v.salary_min, v.salary_max)}</span>}
          <div className="row">
            <Link className="btn" style={{ flex: 1 }} to={`/vacancy/${v.id}/feed`}>
              Лента
            </Link>
            <Link className="btn btn-ghost" style={{ flex: 1 }} to={`/vacancy/${v.id}/list`}>
              Список
            </Link>
            <button className="btn btn-ghost" style={{ flex: 1 }} onClick={() => toggle(v)}>
              {v.is_active ? 'В архив' : 'Вернуть'}
            </button>
          </div>
        </div>
      ))}
    </Screen>
  )
}
