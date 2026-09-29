import { useEffect, useRef, useState } from 'react'
import { mapVacancies } from '../api/map'
import type { MapVacancy } from '../api/types'
import { Screen } from '../components/Screen'
import { formatSalary } from '../components/format'
import { getStoredUser } from '../lib/session'

declare global {
  interface Window {
    ymaps: any
  }
}

const API_KEY: string = import.meta.env.VITE_YANDEX_MAPS_API_KEY ?? ''

function ensureYmaps(): Promise<void> {
  if (window.ymaps) {
    return new Promise((resolve) => window.ymaps.ready(() => resolve()))
  }
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = `https://api-maps.yandex.ru/2.1/?apikey=${API_KEY}&lang=ru_RU`
    script.onload = () => {
      window.ymaps.ready(() => resolve())
    }
    script.onerror = () => reject(new Error('Не удалось загрузить Яндекс.Карты'))
    document.head.appendChild(script)
  })
}

function vacancyBalloon(v: MapVacancy) {
  const body = `${v.company_name}${v.verified ? ' ✓' : ''}<br/>${v.city || ''}`
  const footer = formatSalary(v.salary_min, v.salary_max)
  return {
    balloonContentHeader: v.title,
    balloonContentBody: body,
    balloonContentFooter: footer,
  }
}

export default function MapScreen() {
  const mapRef = useRef<HTMLDivElement>(null)
  const mapInstance = useRef<any>(null)
  const placemarkRefs = useRef<any[]>([])
  const [items, setItems] = useState<MapVacancy[] | null>(null)
  const [error, setError] = useState('')
  const role = getStoredUser<{ role: string } | null>()?.role

  useEffect(() => {
    mapVacancies()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  useEffect(() => {
    let cancelled = false
    if (!API_KEY) return
    if (!items || !items.length) return

    ensureYmaps()
      .then(() => {
        if (cancelled || !mapRef.current || !window.ymaps) return

        if (!mapInstance.current) {
          mapInstance.current = new window.ymaps.Map(mapRef.current, {
            center: [items[0].lat, items[0].lng],
            zoom: items.length === 1 ? 12 : 10,
            controls: ['zoomControl', 'fullscreenControl'],
          })
        }
        const map = mapInstance.current

        placemarkRefs.current.forEach((p) => map.geoObjects.remove(p))
        placemarkRefs.current = []

        items.forEach((v) => {
          if (!v.lat || !v.lng) return
          const placemark = new window.ymaps.Placemark([v.lat, v.lng], vacancyBalloon(v), {
            preset: 'islands#blueDotIcon',
            iconColor: '#5288c1',
          })
          map.geoObjects.add(placemark)
          placemarkRefs.current.push(placemark)
        })

        if (items.length === 1) {
          map.setCenter([items[0].lat, items[0].lng], 12)
        } else if (items.length > 1) {
          map.setBounds(
            [
              [Math.min(...items.map((i) => i.lat)), Math.min(...items.map((i) => i.lng))],
              [Math.max(...items.map((i) => i.lat)), Math.max(...items.map((i) => i.lng))],
            ],
            { checkZoomRange: true, zoomMargin: [80, 80] },
          )
          if (map.getZoom() > 13) {
            map.setZoom(13)
          }
        }
      })
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))

    return () => {
      cancelled = true
    }
  }, [items])

  return (
    <Screen role={role === 'recruiter' ? 'recruiter' : 'candidate'} title="Карта вакансий" icon="map">
      {(!API_KEY || error) && (
        <p className="error-text">{!API_KEY ? 'Карта временно недоступна: не настроен API-ключ Яндекс.Карт' : error}</p>
      )}
      {items === null && !error && <p className="muted">Загрузка…</p>}
      {items !== null && items.length === 0 && (
        <div className="center-note">
          <span>Пока нет вакансий на карте</span>
        </div>
      )}

      <div
        ref={mapRef}
        style={{ height: '62vh', borderRadius: 12, overflow: 'hidden', border: '1px solid var(--stroke)' }}
      />

      {items !== null && items.length > 0 && (
        <p className="muted" style={{ textAlign: 'center', fontSize: 12, margin: 0 }}>
          {items.length} вакансий · Яндекс.Карты
        </p>
      )}
    </Screen>
  )
}
