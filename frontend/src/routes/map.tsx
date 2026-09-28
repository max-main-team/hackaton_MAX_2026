import { useEffect, useRef, useState } from 'react'
import { mapVacancies } from '../api/map'
import type { MapVacancy } from '../api/types'

declare global {
  interface Window {
    ymaps: any
  }
}

let ymapsLoaded = false
let ymapsCallbacks: (() => void)[] = []

function loadYmaps(): Promise<void> {
  if (ymapsLoaded) return Promise.resolve()
  if (ymapsCallbacks.length) {
    return new Promise(resolve => ymapsCallbacks.push(resolve))
  }

  return new Promise(resolve => {
    ymapsCallbacks.push(() => { ymapsLoaded = true; resolve() })
    const script = document.createElement('script')
    script.src = 'https://api-maps.yandex.ru/2.1/?lang=ru_RU'
    script.onload = () => {
      ymapsCallbacks.forEach(f => f())
      ymapsCallbacks = []
    }
    document.head.appendChild(script)
  })
}

export default function MapScreen() {
  const mapRef = useRef<HTMLDivElement>(null)
  const [items, setItems] = useState<MapVacancy[] | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    mapVacancies()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : String(e)))
  }, [])

  useEffect(() => {
    if (!items || !items.length || !mapRef.current) return

    loadYmaps().then(() => {
      if (!window.ymaps || !mapRef.current) return

      const map = new window.ymaps.Map(mapRef.current, {
        center: [59.9386, 30.3141],
        zoom: 11,
        controls: ['zoomControl', 'fullscreenControl']
      })

      items.forEach((v) => {
        if (!v.lat || !v.lng) return
        const placemark = new window.ymaps.Placemark([v.lat, v.lng], {
          balloonContentHeader: v.title,
          balloonContentBody: `${v.company_name}${v.verified ? ' ✓' : ''}<br/>${v.city || ''}`,
          balloonContentFooter: v.salary_min ? `от ${v.salary_min}₽` : ''
        }, {
          preset: 'islands#blueDotIcon',
          iconColor: '#5288c1'
        })
        map.geoObjects.add(placemark)
      })

      if (items.length === 1) {
        map.setCenter([items[0].lat, items[0].lng], 13)
      } else if (items.length > 1) {
        const lats = items.map(v => v.lat)
        const lngs = items.map(v => v.lng)
        map.setBounds([[Math.min(...lats), Math.min(...lngs)], [Math.max(...lats), Math.max(...lngs)]], { checkZoomRange: true })
      }
    })
  }, [items])

  return (
    <main className="page">
      <h1 className="page-title">Карта вакансий</h1>
      <p className="page-sub">Активные вакансии с координатами · Яндекс.Карты</p>
      {error && <p className="error-text">{error}</p>}
      {items === null && <p className="muted">Загрузка…</p>}

      {items !== null && items.length === 0 && <p className="muted">Пока нет вакансий на карте</p>}

      <div ref={mapRef} style={{ height: '70vh', borderRadius: 12, overflow: 'hidden' }} />

      {items !== null && items.length > 0 && (
        <p className="muted" style={{ textAlign: 'center' }}>
          {items.length} вакансий · Яндекс.Карты
        </p>
      )}
    </main>
  )
}
