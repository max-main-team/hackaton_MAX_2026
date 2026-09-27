const API_BASE: string = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    const body = await res.text()
    throw new ApiError(res.status, body || res.statusText)
  }
  return res.json() as Promise<T>
}

export interface AuthUser {
  id: number
  username: string
  first_name: string
  last_name: string
  photo_url: string
  created_at: string
  updated_at: string
}

export const api = {
  auth: (initData: string) =>
    request<AuthUser>('/auth', { method: 'POST', body: JSON.stringify({ initData }) }),

  health: () => request<{ status: string; db: string }>('/health'),
}
