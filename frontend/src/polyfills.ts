// Полифиллы для старых вебвью (MAX на iOS/Android).
// Подключается раньше всех в main.tsx.

interface WithResolvers<T> {
  promise: Promise<T>
  resolve: (value: T | PromiseLike<T>) => void
  reject: (reason?: unknown) => void
}

declare global {
  interface PromiseConstructor {
    withResolvers?: <T>() => WithResolvers<T>
  }
}

if (typeof Promise.withResolvers !== 'function') {
  Promise.withResolvers = function <T>(): WithResolvers<T> {
    let resolve!: (value: T | PromiseLike<T>) => void
    let reject!: (reason?: unknown) => void
    const promise = new Promise<T>((res, rej) => {
      resolve = res
      reject = rej
    })
    return { promise, resolve, reject }
  }
}

if (typeof Object.fromEntries !== 'function') {
  Object.fromEntries = function (entries: Iterable<readonly [PropertyKey, unknown]>): Record<PropertyKey, unknown> {
    const result: Record<PropertyKey, unknown> = {}
    for (const [key, value] of entries) {
      result[key] = value
    }
    return result
  }
}

if (typeof Object.hasOwn !== 'function') {
  Object.hasOwn = function (object: object, property: PropertyKey): boolean {
    return Object.prototype.hasOwnProperty.call(object, property)
  }
}

if (typeof Array.prototype.at !== 'function') {
  Object.defineProperty(Array.prototype, 'at', {
    value: function (this: unknown[], index: number) {
      const len = this.length >>> 0
      const i = Math.trunc(index) || 0
      if (i < 0) return this[len + i]
      return this[i]
    },
    enumerable: false,
    writable: true,
    configurable: true,
  })
}

if (typeof String.prototype.replaceAll !== 'function') {
  Object.defineProperty(String.prototype, 'replaceAll', {
    value: function (this: string, search: string | RegExp, replacement: string | ((s: string) => string)): string {
      if (typeof search === 'string') {
        return this.split(search).join(typeof replacement === 'string' ? replacement : replacement(search))
      }
      return this.replace(search, replacement as string)
    },
    enumerable: false,
    writable: true,
    configurable: true,
  })
}

if (typeof Array.prototype.findLast !== 'function') {
  Object.defineProperty(Array.prototype, 'findLast', {
    value: function <T>(this: T[], predicate: (value: T, index: number, array: T[]) => boolean, thisArg?: unknown): T | undefined {
      for (let i = this.length - 1; i >= 0; i--) {
        if (predicate.call(thisArg, this[i] as T, i, this as unknown as T[])) return this[i] as T
      }
      return undefined
    },
    enumerable: false,
    writable: true,
    configurable: true,
  })
}

if (typeof globalThis.structuredClone !== 'function') {
  globalThis.structuredClone = function structuredClone(value: unknown): unknown {
    if (value === null || typeof value !== 'object') return value
    if (value instanceof Date) return new Date(value.getTime())
    if (value instanceof RegExp) return new RegExp(value.source, value.flags)
    if (value instanceof Map) {
      const copy = new Map()
      value.forEach((v, k) => copy.set(k, structuredClone(v)))
      return copy
    }
    if (value instanceof Set) {
      const copy = new Set()
      value.forEach(v => copy.add(structuredClone(v)))
      return copy
    }
    if (ArrayBuffer.isView(value)) {
      const view = value as unknown as { constructor: { new (buffer: ArrayBufferLike): unknown } }
      const buffer = (value as unknown as { buffer: ArrayBufferLike }).buffer.slice(0)
      return new view.constructor(buffer)
    }
    if (Array.isArray(value)) return value.map(item => structuredClone(item))
    const copy: Record<string, unknown> = {}
    for (const key of Object.keys(value as Record<string, unknown>)) {
      copy[key] = structuredClone((value as Record<string, unknown>)[key])
    }
    return copy
  } as typeof structuredClone
}
