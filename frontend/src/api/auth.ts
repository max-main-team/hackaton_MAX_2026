import { request } from './client'
import type { AuthResponse, MeResponse, ReferralsResponse } from './types'

export function auth(initData: string): Promise<AuthResponse> {
  return request<AuthResponse>('/auth', {
    method: 'POST',
    body: JSON.stringify({ initData }),
  })
}

export function me(): Promise<MeResponse> {
  return request<MeResponse>('/me')
}

export function setRole(role: string, acceptPersonalData: boolean): Promise<MeResponse> {
  return request<MeResponse>('/me/role', {
    method: 'POST',
    body: JSON.stringify({ role, acceptPersonalData }),
  })
}

export function myReferrals(): Promise<ReferralsResponse> {
  return request<ReferralsResponse>('/my/referrals')
}
