import { useEffect, useState } from 'react'
import { MapContainer, Marker, Popup, TileLayer } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { mapVacancies } from '../api/map'
import type { MapVacancy } from '../api/types'
import { Badge } from '../components/Badge'

const SPB: [number, number] = [59.9386, 30.3141]

const icon = L.icon({
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
})

export default function MapScreen() {
  const [items, setItems] = useState<MapVacancy[] | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    mapVacancies()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  return (
    <main className="page">
      <h1 className="page-title">Карта вакансий</h1>
      <p className="page-sub">Активные вакансии с координатами</p>
      {error && <p className="error-text">{error}</p>}
      {items === null && <p className="muted">Загрузка…</p>}

      {items !== null && (
        <div style={{ height: '70vh', borderRadius: 12, overflow: 'hidden' }}>
          <MapContainer center={SPB} zoom={11} style={{ height: '100%', width: '100%' }}>
            <TileLayer
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
              url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
            {items.map((v) => (
              <Marker key={v.id} position={[v.lat, v.lng]} icon={icon}>
                <Popup>
                  <strong>{v.title}</strong>
                  <br />
                  {v.company_name} <Badge verified={v.verified} />
                  <br />
                  {v.city}
                  {v.salary_min !== null && v.salary_max !== null && (
                    <>
                      <br />
                      {v.salary_min}–{v.salary_max} ₽
                    </>
                  )}
                </Popup>
              </Marker>
            ))}
          </MapContainer>
        </div>
      )}

      {items !== null && items.length === 0 && <p className="muted">Пока нет вакансий на карте</p>}
    </main>
  )
}
