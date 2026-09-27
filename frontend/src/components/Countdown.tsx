import { useEffect, useState } from 'react'

interface Props {
  deadlineAt: string
}

export function Countdown({ deadlineAt }: Props) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 60000)
    return () => clearInterval(timer)
  }, [])

  const diffHours = (new Date(deadlineAt).getTime() - now) / 3600000

  if (diffHours <= 0) {
    return <span className="error-text">Просрочено</span>
  }
  const whole = Math.floor(diffHours)
  const minutes = Math.floor((diffHours - whole) * 60)
  return (
    <span className="ok-text">
      Осталось {whole} ч {minutes > 0 ? `${minutes} мин` : ''}
    </span>
  )
}
