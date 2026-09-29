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

if (typeof Array.prototype.at !== 'function') {
  Array.prototype.at = function (index: number) {
    const len = this.length >>> 0
    const i = Math.trunc(index) || 0
    if (i < 0) return this[len + i]
    return this[i]
  }
}
