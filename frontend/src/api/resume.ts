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

export function parseResume(text: string, fileName: string): Promise<{ status: string }> {
  return request<{ status: string }>('/my/resume/parse', {
    method: 'POST',
    body: JSON.stringify({ text, file_name: fileName }),
  })
}

export function parseFile(file: File): Promise<{ status: string }> {
  const body = new FormData()
  body.append('file', file)
  return request<{ status: string }>('/my/resume/parse-file', {
    method: 'POST',
    body,
  })
}
