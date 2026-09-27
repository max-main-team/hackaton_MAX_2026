import type { CandidateItem } from '../api/types'
import { WORK_FORMAT_LABELS } from '../api/types'

interface Props {
  item: CandidateItem
  onAction?: (action: 'invite' | 'skip') => void
}

export function UserCard({ item, onAction }: Props) {
  const skills = item.resume.skills
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <span className="muted">{item.resume.title}</span>
      <div className="row">
        {skills.slice(0, 8).map((s) => (
          <span className="tag" key={s}>
            {s}
          </span>
        ))}
      </div>
      <span className="muted">
        {item.resume.city || '—'} · {WORK_FORMAT_LABELS[item.resume.work_format] ?? item.resume.work_format} ·
        опыт {item.resume.experience_months} мес
      </span>
      {item.ai_comment && <span className="muted">AI: {item.ai_comment}</span>}
      {onAction && (
        <div className="row">
          <button className="btn btn-success" onClick={() => onAction('invite')}>
            Пригласить
          </button>
          <button className="btn btn-secondary" onClick={() => onAction('skip')}>
            Пропустить
          </button>
        </div>
      )}
    </div>
  )
}
