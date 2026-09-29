export function formatSalary(min: number | null, max: number | null): string {
  if (min && max) return `${min.toLocaleString('ru-RU')} – ${max.toLocaleString('ru-RU')} ₽`
  if (min) return `от ${min.toLocaleString('ru-RU')} ₽`
  if (max) return `до ${max.toLocaleString('ru-RU')} ₽`
  return ''
}

export function shortCity(city: string): string {
  if (city === 'Санкт-Петербург') return 'СПб'
  if (city === 'Москва') return 'МСК'
  return city
}

export function experienceLabel(months: number): string {
  const years = Math.round(months / 12)
  return years > 0 ? `${years} лет` : `${months} мес`
}

export function digits(raw: string): string {
  return raw.replace(/[^0-9]/g, '')
}
