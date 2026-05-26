const RECOVERY_KEY = 'edu-schedule-system.chunk-reload'

function currentPath() {
  return `${window.location.pathname}${window.location.search}${window.location.hash}`
}

function buildRecoveryToken(path: string) {
  return `recover:${path}`
}

function canRecover(path: string) {
  if (typeof window === 'undefined') {
    return false
  }
  const token = buildRecoveryToken(path)
  const last = window.sessionStorage.getItem(RECOVERY_KEY)
  if (last === token) {
    window.sessionStorage.removeItem(RECOVERY_KEY)
    return false
  }
  window.sessionStorage.setItem(RECOVERY_KEY, token)
  return true
}

export function installChunkLoadRecovery() {
  if (typeof window === 'undefined') {
    return
  }

  window.addEventListener('error', (event) => {
    const target = event.target
    const source = target instanceof HTMLScriptElement
      ? target.src
      : target instanceof HTMLLinkElement
        ? target.href
        : ''
    const isRecoverableAsset = /\/assets\/.+\.(js|css)$/i.test(source)
      || /\/(src|@vite|node_modules\/\.vite)\//.test(source)
    if (!isRecoverableAsset) {
      return
    }

    if (!canRecover(currentPath())) {
      return
    }

    window.location.replace(currentPath())
  }, true)

  window.addEventListener('unhandledrejection', (event) => {
    const reason = event.reason
    const message = typeof reason?.message === 'string' ? reason.message : String(reason || '')
    const chunkLoadFailed = /Failed to fetch dynamically imported module|Importing a module script failed|Loading chunk [\w-]+ failed|Outdated Optimize Dep|fetch dynamically imported module|Failed to load url \/@vite\//i.test(message)
    if (!chunkLoadFailed) {
      return
    }

    if (!canRecover(currentPath())) {
      return
    }

    window.location.replace(currentPath())
  })
}
