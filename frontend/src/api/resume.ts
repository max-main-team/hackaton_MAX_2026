import { request } from './client'
import type { Resume, ResumeInput } from './types'

export function getResume(): Promise<Resume> {
  return request<Resume>('/my/resume')
}

export function saveResume(input: ResumeInput): Promise<Resume> {
  return request<Resume>('/my/resume', { method: 'PUT', body: JSON.stringify(input) })
}

export function confirmActivity(active: boolean): Promise<Resume> {
  return request<Resume>('/my/resume/confirm-activity', {
    method: 'POST',
    body: JSON.stringify({ active }),
  })
}
