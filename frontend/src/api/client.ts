import { authHeaders, clearSession } from '../lib/session'

export const API_BASE: string = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

const DEFAULT_TIMEOUT_MS = 30_000

export async function request<T>(path: string, init?: RequestInit, timeoutMs = DEFAULT_TIMEOUT_MS): Promise<T> {
  const headers = new Headers(init?.headers)
  for (const [name, value] of Object.entries(authHeaders())) {
    if (!headers.has(name)) headers.set(name, value)
  }

  // Браузер сам добавляет boundary для FormData. Если выставить здесь JSON,
  // multipart ломается (особенно заметно в iOS WebView).
  if (typeof init?.body === 'string' && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const controller = new AbortController()
  const externalSignal = init?.signal
  const abort = () => controller.abort()
  if (externalSignal?.aborted) abort()
  else externalSignal?.addEventListener('abort', abort, { once: true })
  const timeout = window.setTimeout(abort, timeoutMs)

  let res: Response
  try {
    res = await fetch(`${API_BASE}${path}`, {
      ...init,
      headers,
      signal: controller.signal,
    })
  } catch (error) {
    if (controller.signal.aborted && !externalSignal?.aborted) {
      throw new ApiError(0, 'Сервер не ответил вовремя — попробуйте ещё раз')
    }
    throw error
  } finally {
    window.clearTimeout(timeout)
    externalSignal?.removeEventListener('abort', abort)
  }

  if (res.status === 401) {
    clearSession()
    if (window.location.pathname !== '/') {
      window.location.assign('/')
    }
    throw new ApiError(401, 'unauthorized')
  }
  if (!res.ok) {
    if (res.status >= 500) {
      throw new ApiError(res.status, 'Ошибка сервера — попробуйте позже')
    }
    let message = res.statusText
    try {
      const body = await res.json()
      if (body && typeof body.message === 'string') message = body.message
    } catch {
      /* not json */
    }
    throw new ApiError(res.status, message)
  }
  return res.json() as Promise<T>
}
