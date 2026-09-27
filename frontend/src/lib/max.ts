/**
 * Типы и утилиты для MAX Bridge — глобального объекта window.WebApp.
 * Скрипт бриджа подключается в index.html.
 */

export interface MaxUser {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  photo_url?: string
}

export interface MaxWebApp {
  initData: string
  initDataUnsafe: {
    user?: MaxUser
    start_param?: string
  }
  version: string
  platform: string
  colorScheme: 'light' | 'dark'
  themeParams: Record<string, string>
  isExpanded: boolean
  ready: () => void
  expand: () => void
  close: () => void
  sendData: (data: string) => void
  onEvent: (event: string, handler: (...args: unknown[]) => void) => void
  offEvent: (event: string, handler: (...args: unknown[]) => void) => void
  HapticFeedback?: {
    impactOccurred: (style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft') => void
    notificationOccurred: (type: 'error' | 'success' | 'warning') => void
    selectionChanged: () => void
  }
}

export function getWebApp(): MaxWebApp | null {
  return (window as unknown as { WebApp?: MaxWebApp }).WebApp ?? null
}

export function isInsideMax(): boolean {
  return getWebApp() !== null
}

export function getInitData(): string {
  return getWebApp()?.initData ?? ''
}
