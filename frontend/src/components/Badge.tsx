export function Badge({ verified }: { verified: boolean }) {
  if (!verified) return null
  return <span className="badge">✓ Верифицирована</span>
}
