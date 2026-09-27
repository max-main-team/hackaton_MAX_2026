import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { createCompany, myCompanies } from '../../api/company'
import type { Company } from '../../api/types'

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

  const load = () => {
    myCompanies()
      .then(setCompanies)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }

  useEffect(load, [])

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
      <main className="page">
        <p className="muted">Загрузка…</p>
      </main>
    )
  }

  const current = companies[0]

  return (
    <main className="page">
      <h1 className="page-title">Моя компания</h1>

      {current ? (
        <>
          <div className="card">
            <div className="row" style={{ justifyContent: 'space-between' }}>
              <strong>{current.name}</strong>
              {current.verified && <span className="badge">✓ Верифицирована</span>}
            </div>
            {current.description && <span className="muted">{current.description}</span>}
            {current.address && <span className="muted">{current.address}</span>}
            {current.website && (
              <a className="muted" href={current.website} target="_blank" rel="noreferrer">
                {current.website}
              </a>
            )}
          </div>
          <Link className="btn" to="/company/vacancies" style={{ textAlign: 'center' }}>
            Мои вакансии
          </Link>
        </>
      ) : (
        <>
          <div className="card">
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
              <input className="input" value={form.website} onChange={(e) => setForm({ ...form, website: e.target.value })} />
            </div>
            <div className="field">
              <label>Адрес</label>
              <input className="input" value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} />
            </div>
            <div className="field">
              <label>Ваша позиция</label>
              <select className="select" value={form.position} onChange={(e) => setForm({ ...form, position: e.target.value })}>
                <option value="owner">Директор / владелец</option>
                <option value="hr">HR</option>
                <option value="employee">Сотрудник</option>
              </select>
            </div>
            {error && <p className="error-text">{error}</p>}
            <button className="btn" onClick={create} disabled={!form.name.trim()}>
              Создать компанию
            </button>
          </div>
          <p className="muted">
            Верификация компании появится здесь же — понадобится токен бота компании.
          </p>
        </>
      )}
    </main>
  )
}
