const RUNTIME_ERROR_KEY = 'edu-schedule-system.runtime-error'

export interface RuntimeErrorSnapshot {
  scope: 'router' | 'vue' | 'window'
  route: string
  message: string
  time: string
}

export function saveRuntimeError(snapshot: RuntimeErrorSnapshot) {
  if (typeof window === 'undefined') {
    return
  }
  window.sessionStorage.setItem(RUNTIME_ERROR_KEY, JSON.stringify(snapshot))
}

export function readRuntimeError(): RuntimeErrorSnapshot | null {
  if (typeof window === 'undefined') {
    return null
  }
  const raw = window.sessionStorage.getItem(RUNTIME_ERROR_KEY)
  if (!raw) {
    return null
  }
  try {
    return JSON.parse(raw) as RuntimeErrorSnapshot
  } catch {
    return null
  }
}

export function clearRuntimeError() {
  if (typeof window === 'undefined') {
    return
  }
  window.sessionStorage.removeItem(RUNTIME_ERROR_KEY)
}

export function summarizeUnknownError(error: unknown) {
  if (error instanceof Error) {
    return error.message || error.name || 'unknown error'
  }
  if (typeof error === 'string') {
    return error
  }
  try {
    return JSON.stringify(error)
  } catch {
    return String(error)
  }
}
