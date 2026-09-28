import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { createCompany, myCompanies } from '../../api/company'
import type { Company } from '../../api/types'
import { Screen } from '../../components/Screen'
import { VerifiedBadge } from '../../components/Badge'

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
      <Screen role="recruiter">
        <p className="muted">Загрузка…</p>
      </Screen>
    )
  }

  const current = companies[0]

  return (
    <Screen role="recruiter" title="Моя компания" icon="briefcase">
      {error && <p className="error-text">{error}</p>}

      {current ? (
        <>
          <div className="card gap-sm">
            <div className="row between" style={{ flexWrap: 'nowrap' }}>
              <span className="avatar-sq">{current.name.charAt(0).toUpperCase()}</span>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: 1, minWidth: 0 }}>
                <div className="row" style={{ flexWrap: 'nowrap', gap: 4 }}>
                  <strong style={{ fontSize: 15, overflow: 'hidden', textOverflow: 'ellipsis' }}>{current.name}</strong>
                  <VerifiedBadge verified={current.verified} />
                </div>
                {current.website && (
                  <a className="muted" style={{ fontSize: 12 }} href={current.website} target="_blank" rel="noreferrer">
                    {current.website}
                  </a>
                )}
              </div>
            </div>
            {current.description && <span className="muted">{current.description}</span>}
            {current.address && <span className="muted">{current.address}</span>}
          </div>
          <Link className="btn" to="/company/vacancies">
            Мои вакансии
          </Link>
        </>
      ) : (
        <>
          <div className="form-section">
            <p className="section-label">Новая компания</p>
            <div className="fieldset">
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
                <input
                  className="input"
                  value={form.website}
                  onChange={(e) => setForm({ ...form, website: e.target.value })}
                />
              </div>
              <div className="field">
                <label>Адрес</label>
                <input
                  className="input"
                  value={form.address}
                  onChange={(e) => setForm({ ...form, address: e.target.value })}
                />
              </div>
              <div className="field">
                <label>Ваша позиция</label>
                <select
                  className="select"
                  value={form.position}
                  onChange={(e) => setForm({ ...form, position: e.target.value })}
                >
                  <option value="owner">Директор / владелец</option>
                  <option value="hr">HR</option>
                  <option value="employee">Сотрудник</option>
                </select>
              </div>
            </div>
          </div>

          <button className="btn btn-xl" onClick={create} disabled={!form.name.trim()}>
            Создать компанию
          </button>

          <p className="muted" style={{ fontSize: 12, margin: 0 }}>
            Верификация компании появится здесь же — понадобится токен бота компании.
          </p>
        </>
      )}
    </Screen>
  )
}
