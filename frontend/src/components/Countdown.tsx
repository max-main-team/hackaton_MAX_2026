import { useEffect, useState } from 'react'
import { Icon } from './Icon'

interface Props {
  deadlineAt: string
  hoursLeft?: number
}

function format(hoursDiff: number): { text: string; warn: boolean } {
  const whole = Math.floor(hoursDiff)
  const minutes = Math.floor((hoursDiff - whole) * 60)
  const text = `Осталось ${whole} ч${minutes > 0 ? ` ${minutes} мин` : ''}`
  return { text, warn: hoursDiff < 2 }
}

export function Countdown({ deadlineAt, hoursLeft }: Props) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 60000)
    return () => clearInterval(timer)
  }, [])

  const diffHours = hoursLeft ?? (new Date(deadlineAt).getTime() - now) / 3600000

  if (diffHours <= 0) {
    return <span className="timer overdue">Просрочено</span>
  }
  const { text, warn } = format(diffHours)
  return (
    <span className={`timer${warn ? ' warn' : ''}`}>
      {warn ? <Icon name="alertCircle" size={12} /> : <Icon name="clock" size={12} />}
      {text}
    </span>
  )
}
