import { request } from './client'
import type {
  CandidatesResponse,
  Invitation,
  MatchResult,
  RecruiterActionResponse,
} from './types'

export function candidates(
  vacancyId: number,
  mode: 'list' | 'feed',
  limit = 20,
  offset = 0,
): Promise<CandidatesResponse> {
  const qs = new URLSearchParams({ mode, limit: String(limit), offset: String(offset) })
  return request<CandidatesResponse>(`/vacancies/${vacancyId}/candidates?${qs}`)
}

export function candidateAction(
  vacancyId: number,
  candidateUserId: number,
  action: 'invite' | 'skip',
): Promise<RecruiterActionResponse> {
  return request<RecruiterActionResponse>(
    `/vacancies/${vacancyId}/candidates/${candidateUserId}/action`,
    { method: 'POST', body: JSON.stringify({ action }) },
  )
}

export function invitations(): Promise<Invitation[]> {
  return request<Invitation[]>('/my/invitations')
}

export function respondInvitation(
  id: number,
  response: 'accept' | 'decline',
): Promise<MatchResult> {
  return request<MatchResult>(`/invitations/${id}/respond`, {
    method: 'POST',
    body: JSON.stringify({ response }),
  })
}
