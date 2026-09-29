import { useEffect, useRef, useState } from 'react'
import { mapVacancies } from '../api/map'
import type { MapVacancy } from '../api/types'
import { Screen } from '../components/Screen'
import { formatSalary, shortCity } from '../components/format'
import { VerifiedBadge } from '../components/Badge'
import { getStoredUser } from '../lib/session'

declare global {
  interface Window {
    ymaps3?: any
  }
}

const API_KEY: string = import.meta.env.VITE_YANDEX_MAPS_API_KEY ?? ''

function ensureYmaps3(): Promise<void> {
  if (window.ymaps3) return Promise.resolve()
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = `https://api-maps.yandex.ru/v3/?apikey=${API_KEY}&lang=ru_RU`
    script.onload = () => {
      const wait = () => {
        if (window.ymaps3) {
          window.ymaps3.ready.then(() => resolve()).catch(reject)
        } else {
          setTimeout(wait, 200)
        }
      }
      wait()
    }
    script.onerror = () => reject(new Error('Не удалось загрузить Яндекс.Карты'))
    document.head.appendChild(script)
  })
}

function boundsOf(items: MapVacancy[]): [[number, number], [number, number]] {
  const lngs = items.map((v) => v.lng)
  const lats = items.map((v) => v.lat)
  return [
    [Math.min(...lngs), Math.min(...lats)],
    [Math.max(...lngs), Math.max(...lats)],
  ]
}

export default function MapScreen() {
  const mapRef = useRef<HTMLDivElement>(null)
  const mapInstance = useRef<any>(null)
  const [items, setItems] = useState<MapVacancy[] | null>(null)
  const [selected, setSelected] = useState<MapVacancy | null>(null)
  const [error, setError] = useState('')
  const role = getStoredUser<{ role: string } | null>()?.role

  useEffect(() => {
    mapVacancies()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  useEffect(() => {
    let cancelled = false
    if (!API_KEY) {
      setError('Карта временно недоступна: не настроен API-ключ Яндекс.Карт')
      return
    }
    if (!items || !items.length) return

    ensureYmaps3()
      .then(() => {
        if (cancelled || !mapRef.current || !window.ymaps3) return
        const { YMap, YMapDefaultSchemeLayer, YMapDefaultFeaturesLayer, YMapMarker } = window.ymaps3

        if (!mapInstance.current) {
          mapInstance.current = new YMap(mapRef.current, {
            location: { center: [items[0].lng, items[0].lat], zoom: items.length === 1 ? 13 : 11 },
            theme: 'dark',
          })
          mapInstance.current.addChild(new YMapDefaultSchemeLayer({}))
          mapInstance.current.addChild(new YMapDefaultFeaturesLayer({}))
        }
        const map = mapInstance.current

        items.forEach((v) => {
          if (!v.lat || !v.lng) return
          const el = document.createElement('button')
          el.className = 'map-marker'
          el.type = 'button'
          el.innerHTML = `<span>${v.title}</span>`
          el.addEventListener('click', (e) => {
            e.stopPropagation()
            setSelected(v)
          })
          map.addChild(new YMapMarker({ coordinates: [v.lng, v.lat] }, el))
        })

        if (items.length > 1) {
          map.update({ location: { bounds: boundsOf(items) } })
        }
      })
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))

    return () => {
      cancelled = true
    }
  }, [items])

  return (
    <Screen role={role === 'recruiter' ? 'recruiter' : 'candidate'} title="Карта вакансий" icon="map">
      {error && <p className="error-text">{error}</p>}
      {items === null && !error && <p className="muted">Загрузка…</p>}
      {items !== null && items.length === 0 && (
        <div className="center-note">
          <span>Пока нет вакансий на карте</span>
        </div>
      )}

      <div
        ref={mapRef}
        style={{ height: '58vh', borderRadius: 12, overflow: 'hidden', border: '1px solid var(--stroke)' }}
      />

      {selected && (
        <div className="card gap-sm" style={{ padding: 14 }}>
          <div className="row between" style={{ flexWrap: 'nowrap' }}>
            <div className="row" style={{ flexWrap: 'nowrap', gap: 4, minWidth: 0 }}>
              <strong style={{ fontSize: 14, overflow: 'hidden', textOverflow: 'ellipsis' }}>{selected.company_name}</strong>
              <VerifiedBadge verified={selected.verified} />
            </div>
            <button className="link-btn" onClick={() => setSelected(null)}>
              ✕
            </button>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <span style={{ fontSize: 16, fontWeight: 700 }}>{selected.title}</span>
            <span className="muted" style={{ fontSize: 12 }}>
              {selected.city ? `${shortCity(selected.city)} · ` : ''}
              {formatSalary(selected.salary_min, selected.salary_max)}
            </span>
          </div>
        </div>
      )}

      {items !== null && items.length > 0 && (
        <p className="muted" style={{ textAlign: 'center', fontSize: 12, margin: 0 }}>
          {items.length} вакансий · Яндекс.Карты
        </p>
      )}
    </Screen>
  )
}
