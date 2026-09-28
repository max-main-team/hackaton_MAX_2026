/**
 * Типы MAX Bridge — window.WebApp (https://dev.max.ru/docs/webapps/bridge).
 * Выжимка: docs/max/bridge.md.
 * Инициализация не требуется: объект создаётся клиентом MAX при запуске.
 * Вне MAX объект отсутствует — всегда проверяйте getWebApp() !== null.
 */

export interface MaxUser {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  language_code?: string
  photo_url?: string
}

export interface MaxWebApp {
  /** Подписанная строка initData для валидации на бэкенде */
  initData: string
  /** Данные из initData в виде объекта. Для валидации непригоден */
  initDataUnsafe: {
    query_id?: string
    ip?: string
    auth_date?: number
    hash?: string
    user?: MaxUser
    chat?: { id: number; type: 'DIALOG' | 'CHAT' | 'CHANNEL' }
    start_param?: string
  }
  /** ios | android | desktop | web */
  platform: string
  version: string
  deviceName?: string
  getLaunchContext?: () => Promise<{ entryPoint: 'tabbar' | 'default' }>
  getViewportSize?: () => Promise<{ height: string; width: string }>
  requestContact?: () => Promise<{ phone: string; authDate: string; hash: string }>
  enableClosingConfirmation?: () => void
  disableClosingConfirmation?: () => void
  openLink?: (url: string) => void
  openMaxLink?: (url: string) => void
  downloadFile?: (url: string, fileName: string) => void
  shareContent?: (params: { text?: string; link?: string }) => void
  shareMaxContent?: (
    params: { text?: string; link?: string } | { mid: string; chatType: 'DIALOG' | 'CHAT' },
  ) => void
  openCodeReader?: (fileSelect?: boolean) => Promise<string>
  BackButton?: {
    show: () => void
    hide: () => void
    isVisible: boolean
    onClick: (cb: () => void) => void
    offClick: (cb: () => void) => void
  }
  DeviceStorage?: {
    setItem: (key: string, value: string) => void
    getItem: (key: string) => string
    removeItem: (key: string) => void
    clear: () => void
  }
  SecureStorage?: {
    setItem: (key: string, value: string) => void
    getItem: (key: string) => string
    removeItem: (key: string) => void
    clear: () => void
  }
  HapticFeedback?: {
    impactOccurred: (style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft', disableFallback?: boolean) => void
    notificationOccurred: (type: 'error' | 'success' | 'warning', disableFallback?: boolean) => void
    selectionChanged: (disableFallback?: boolean) => void
  }
}

export function getWebApp(): MaxWebApp | null {
  return (window as unknown as { WebApp?: MaxWebApp }).WebApp ?? null
}

/** true, если мини-приложение открыто внутри MAX */
export function isInsideMax(): boolean {
  return getWebApp() !== null
}

/** initData для POST /auth (пустая строка вне MAX) */
export function getInitData(): string {
  return getWebApp()?.initData ?? ''
}

/** Безопасный вызов метода Bridge: метод может отсутствовать в версии клиента */
export function callBridge(fn: (app: MaxWebApp) => void): void {
  const app = getWebApp()
  if (!app) return
  try {
    fn(app)
  } catch {
    /* Bridge не поддерживает метод в этой версии клиента */
  }
}
