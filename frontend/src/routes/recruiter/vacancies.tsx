import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { companyVacancies, myCompanies, updateVacancy } from '../../api/company'
import type { Company, Vacancy } from '../../api/types'
import { WORK_FORMAT_LABELS } from '../../api/types'

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
      <main className="page">
        <p className="muted">Загрузка…</p>
      </main>
    )
  }

  return (
    <main className="page">
      <Link className="back" to="/company">
        ← Компания
      </Link>
      <h1 className="page-title">Вакансии{company ? ` · ${company.name}` : ''}</h1>

      {error && <p className="error-text">{error}</p>}

      <Link className="btn" to="/company/vacancies/new" style={{ textAlign: 'center' }}>
        + Новая вакансия
      </Link>

      {vacancies.length === 0 && <p className="muted">Пока нет вакансий</p>}

      {vacancies.map((v) => (
        <div className="list-item" key={v.id}>
          <div className="row" style={{ justifyContent: 'space-between' }}>
            <strong>{v.title}</strong>
            <span className="badge" style={v.is_active ? {} : { background: '#5c2a2a' }}>
              {v.is_active ? 'Активна' : 'Архив'}
            </span>
          </div>
          <span className="muted">
            {WORK_FORMAT_LABELS[v.work_format] ?? v.work_format} · {v.city || '—'} · TTL {v.response_ttl_hours}ч
          </span>
          <div className="row">
            <Link className="btn" style={{ padding: '8px 14px' }} to={`/vacancy/${v.id}/feed`}>
              Лента
            </Link>
            <Link className="btn btn-secondary" style={{ padding: '8px 14px' }} to={`/vacancy/${v.id}/list`}>
              Список
            </Link>
            <button className="btn btn-secondary" style={{ padding: '8px 14px' }} onClick={() => toggle(v)}>
              {v.is_active ? 'В архив' : 'Вернуть'}
            </button>
          </div>
        </div>
      ))}
    </main>
  )
}
