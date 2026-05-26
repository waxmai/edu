import { defineStore } from 'pinia'
import {
  fetchAuthSessions,
  fetchCurrentUser,
  logout as logoutRequest,
  logoutAllSessions,
  refreshToken as refreshTokenRequest,
  revokeAuthSession,
  type AuthSessionInfo,
  type AuthUser,
} from '@/api/auth'

const TOKEN_STORAGE_KEY = 'edu-schedule-system.token'
const REFRESH_TOKEN_STORAGE_KEY = 'edu-schedule-system.refresh-token'
const USER_STORAGE_KEY = 'edu-schedule-system.current-user'
const DEVICE_NAME_STORAGE_KEY = 'edu-schedule-system.device-name'
const FORCE_PASSWORD_CHANGE_KEY = 'edu-schedule-system.force-password-change'
const PLATFORM_VIEW_AS_KEY = 'edu-schedule-system.platform-view-as'

interface PlatformViewAsState {
  organizationId?: number
  organizationName?: string
  campusId?: number
  campusName?: string
}

function readToken() {
  if (typeof window === 'undefined') {
    return ''
  }
  return window.localStorage.getItem(TOKEN_STORAGE_KEY) || ''
}

function readRefreshToken() {
  if (typeof window === 'undefined') {
    return ''
  }
  return window.localStorage.getItem(REFRESH_TOKEN_STORAGE_KEY) || ''
}

function readCurrentUser(): AuthUser | null {
  if (typeof window === 'undefined') {
    return null
  }
  const raw = window.localStorage.getItem(USER_STORAGE_KEY)
  if (!raw) {
    return null
  }
  try {
    return JSON.parse(raw) as AuthUser
  } catch {
    return null
  }
}

function readDeviceName() {
  if (typeof window === 'undefined') {
    return 'Web Console'
  }
  return window.localStorage.getItem(DEVICE_NAME_STORAGE_KEY) || 'Web Console'
}

function readForcePasswordChangeActive() {
  if (typeof window === 'undefined') {
    return false
  }
  return window.sessionStorage.getItem(FORCE_PASSWORD_CHANGE_KEY) === '1'
}

function readPlatformViewAs(): PlatformViewAsState | null {
  if (typeof window === 'undefined') {
    return null
  }
  const raw = window.sessionStorage.getItem(PLATFORM_VIEW_AS_KEY)
  if (!raw) {
    return null
  }
  try {
    return JSON.parse(raw) as PlatformViewAsState
  } catch {
    return null
  }
}

function buildPermissionSet(user: AuthUser | null) {
  return new Set(user?.permissions || [])
}

function buildMenuPermissionSet(user: AuthUser | null) {
  return new Set(user?.menuPermissions || [])
}

function buildFeatureFlagSet(user: AuthUser | null) {
  return new Set(user?.featureFlags || [])
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: readToken(),
    refreshToken: readRefreshToken(),
    currentUser: readCurrentUser(),
    deviceName: readDeviceName(),
    sessions: [] as AuthSessionInfo[],
    bootstrapped: false,
    forcePasswordChangeActive: readForcePasswordChangeActive(),
    platformViewAs: readPlatformViewAs() as PlatformViewAsState | null,
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.token),
    isAdmin: (state) => ['platform_admin', 'org_admin', 'campus_admin'].includes(state.currentUser?.roleCode || ''),
    isTeacher: (state) => state.currentUser?.roleCode === 'teacher',
    mustChangePassword: (state) => Boolean(state.currentUser?.mustChangePassword),
    permissionSet: (state) => buildPermissionSet(state.currentUser),
    menuPermissionSet: (state) => buildMenuPermissionSet(state.currentUser),
    featureFlagSet: (state) => buildFeatureFlagSet(state.currentUser),
  },
  actions: {
    hasAnyRole(roles: string[] = []) {
      if (roles.length === 0) {
        return true
      }
      const currentRole = this.currentUser?.roleCode || ''
      return roles.includes(currentRole)
    },
    hasPermission(permission?: string) {
      if (!permission) {
        return true
      }
      return this.permissionSet.has(permission)
    },
    hasMenuPermission(permission?: string) {
      if (!permission) {
        return true
      }
      return this.menuPermissionSet.has(permission)
    },
    hasFeatureFlag(flag?: string) {
      if (!flag) {
        return true
      }
      return this.featureFlagSet.has(flag)
    },
    setPlatformViewAs(payload: PlatformViewAsState | null) {
      this.platformViewAs = payload
      if (typeof window !== 'undefined') {
        if (payload) {
          window.sessionStorage.setItem(PLATFORM_VIEW_AS_KEY, JSON.stringify(payload))
        } else {
          window.sessionStorage.removeItem(PLATFORM_VIEW_AS_KEY)
        }
      }
    },
    hasPlatformViewAsScope() {
      return this.currentUser?.roleCode === 'platform_admin' && Boolean(this.platformViewAs?.organizationId)
    },
    setToken(token: string) {
      this.token = token
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(TOKEN_STORAGE_KEY, token)
      }
    },
    setRefreshToken(token: string) {
      this.refreshToken = token
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(REFRESH_TOKEN_STORAGE_KEY, token)
      }
    },
    setCurrentUser(user: AuthUser | null) {
      this.currentUser = user
      if (typeof window !== 'undefined') {
        if (user) {
          window.localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user))
        } else {
          window.localStorage.removeItem(USER_STORAGE_KEY)
        }
      }
      console.info('[auth] setCurrentUser', { mustChangePassword: Boolean(user?.mustChangePassword) })
    },
    setDeviceName(name: string) {
      this.deviceName = name || 'Web Console'
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(DEVICE_NAME_STORAGE_KEY, this.deviceName)
      }
    },
    clearAuth() {
      this.token = ''
      this.refreshToken = ''
      this.currentUser = null
      this.sessions = []
      this.bootstrapped = true
      this.forcePasswordChangeActive = false
      this.platformViewAs = null
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem(TOKEN_STORAGE_KEY)
        window.localStorage.removeItem(REFRESH_TOKEN_STORAGE_KEY)
        window.localStorage.removeItem(USER_STORAGE_KEY)
        window.sessionStorage.removeItem(FORCE_PASSWORD_CHANGE_KEY)
        window.sessionStorage.removeItem(PLATFORM_VIEW_AS_KEY)
      }
    },
    activateForcePasswordChange() {
      this.forcePasswordChangeActive = true
      if (typeof window !== 'undefined') {
        window.sessionStorage.setItem(FORCE_PASSWORD_CHANGE_KEY, '1')
      }
      console.info('[auth] activateForcePasswordChange', { active: this.forcePasswordChangeActive })
    },
    deactivateForcePasswordChange() {
      this.forcePasswordChangeActive = false
      if (typeof window !== 'undefined') {
        window.sessionStorage.removeItem(FORCE_PASSWORD_CHANGE_KEY)
      }
      console.info('[auth] deactivateForcePasswordChange', { active: this.forcePasswordChangeActive })
    },
    async tryRefresh() {
      if (!this.refreshToken) {
        return false
      }
      try {
        const response = await refreshTokenRequest({ refreshToken: this.refreshToken })
        const payload = response.data?.data
        if (!payload?.token || !payload?.refreshToken || !payload?.userInfo) {
          throw new Error('refresh payload missing')
        }
        this.setToken(payload.token)
        this.setRefreshToken(payload.refreshToken)
        this.setCurrentUser(payload.userInfo)
        if (payload.userInfo.mustChangePassword) {
          this.activateForcePasswordChange()
        } else {
          this.deactivateForcePasswordChange()
        }
        if (payload.session) {
          this.upsertSession(payload.session)
        }
        return true
      } catch {
        this.clearAuth()
        return false
      }
    },
    async bootstrap() {
      if (!this.token) {
        this.bootstrapped = true
        return
      }
      try {
        const response = await fetchCurrentUser()
        const currentUser = response.data.data
        this.setCurrentUser(currentUser)
        if (currentUser?.mustChangePassword) {
          this.activateForcePasswordChange()
        } else {
          this.deactivateForcePasswordChange()
        }
      } catch {
        const refreshed = await this.tryRefresh()
        if (!refreshed) {
          this.clearAuth()
          return
        }
      } finally {
        this.bootstrapped = true
      }
    },
    async logout() {
      try {
        if (this.token) {
          await logoutRequest()
        }
      } catch {
        // ignore logout transport failure
      } finally {
        this.clearAuth()
      }
    },
    async loadSessions() {
      if (!this.isAuthenticated) {
        this.sessions = []
        return []
      }
      const response = await fetchAuthSessions()
      this.sessions = response.data?.data?.items || []
      return this.sessions
    },
    upsertSession(session: AuthSessionInfo) {
      if (!session?.id) {
        return
      }
      const index = this.sessions.findIndex((item) => item.id === session.id)
      if (index >= 0) {
        this.sessions[index] = session
      } else {
        this.sessions.unshift(session)
      }
    },
    async revokeSession(sessionId: string) {
      await revokeAuthSession(sessionId)
      this.sessions = this.sessions.filter((item) => item.id !== sessionId)
    },
    async logoutAll() {
      await logoutAllSessions()
      this.clearAuth()
    },
  },
})
