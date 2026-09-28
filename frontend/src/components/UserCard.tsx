import type { CandidateItem } from '../api/types'
import { WORK_FORMAT_LABELS } from '../api/types'
import { Icon } from './Icon'

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
      <span className="muted">
        {item.resume.city || '—'} · {WORK_FORMAT_LABELS[item.resume.work_format] ?? item.resume.work_format} · опыт{' '}
        {item.resume.experience_months} мес
      </span>
      <div className="stack-tags">
        {skills.slice(0, 8).map((s) => (
          <span className="chip" key={s}>
            {s}
          </span>
        ))}
      </div>
      {item.ai_comment && (
        <div className="ai-block">
          <span className="ai-label">
            <Icon name="zap" size={14} />
            AI резюме
          </span>
          <p>{item.ai_comment}</p>
        </div>
      )}
      {onAction && (
        <div className="stack-actions" style={{ padding: '8px 0 0' }}>
          <button className="btn btn-ghost" onClick={() => onAction('skip')}>
            Пропустить
          </button>
          <button className="btn" onClick={() => onAction('invite')}>
            Пригласить
          </button>
        </div>
      )}
    </div>
  )
}
