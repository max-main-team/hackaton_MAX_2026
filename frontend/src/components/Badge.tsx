import { Icon } from './Icon'

export function VerifiedBadge({ verified }: { verified: boolean }) {
  if (!verified) return null
  return (
    <span className="verified-dot" title="Верифицирована">
      <Icon name="check" size={8} />
    </span>
  )
}
