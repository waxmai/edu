import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface RefreshPayload {
  refreshToken: string
}

export interface AuthSessionInfo {
  id: string
  deviceId?: string
  deviceName?: string
  clientIP?: string
  userAgent?: string
  status?: string
  current?: boolean
  suspicious?: boolean
  organizationId?: number
  campusId?: number
  roleCode?: string
  riskFlags?: string[]
  lastSeenAt?: string
  lastRefreshedAt?: string
  revokedAt?: string
  revokedReason?: string
  createdAt?: string
  updatedAt?: string
}

export interface AuthUser {
  id: number
  username: string
  realName: string
  roleCode: string
  organizationId?: number
  campusId?: number
  organizationName?: string
  campusName?: string
  subscriptionStatus?: string
  featureFlags?: string[]
  phone?: string
  email?: string
  status?: string
  mustChangePassword?: boolean
  remark?: string
  permissions?: string[]
  menuPermissions?: string[]
  dataScope?: string
}

export interface LoginResponse {
  token: string
  refreshToken: string
  tokenType: string
  expiresIn: number
  userInfo: AuthUser
  session?: AuthSessionInfo
}

export interface SessionListResponse {
  items: AuthSessionInfo[]
  total: number
}

export interface PasswordRecoveryStartPayload {
  username: string
  channel: string
  secondFactorChannel?: string
}

export interface PasswordRecoveryStartResponse {
  accepted: boolean
  challengeId?: string
  expiresIn?: number
  primaryChannel?: string
  primaryTargetMasked?: string
  secondFactorRequired?: boolean
  secondFactorChannel?: string
  secondTargetMasked?: string
  riskFlags?: string[]
  message?: string
}

export interface PasswordRecoveryResetPayload {
  challengeId: string
  verificationCode: string
  secondFactorCode?: string
  newPassword: string
}

export interface PasswordRecoveryResetResponse {
  success: boolean
  revokedSessionCount: number
  requiresFreshLogin: boolean
  recoveredAt: string
  recoveryChallengeId: string
}

export interface LoginPayload {
  username: string
  password: string
}

export function login(payload: LoginPayload) {
  return request.post<ApiResponse<LoginResponse>>('/auth/login', payload)
}

export function refreshToken(payload: RefreshPayload) {
  return request.post<ApiResponse<LoginResponse>>('/auth/refresh', payload)
}

export function fetchCurrentUser() {
  return request.get<ApiResponse<AuthUser>>('/auth/me')
}

export interface ChangePasswordPayload {
  oldPassword: string
  newPassword: string
}

export function changePassword(payload: ChangePasswordPayload) {
  return request.post<ApiResponse<LoginResponse>>('/auth/change-password', payload)
}

export function logout() {
  return request.post<ApiResponse<{ success: boolean }>>('/auth/logout')
}

export function fetchAuthSessions() {
  return request.get<ApiResponse<SessionListResponse>>('/auth/sessions')
}

export function revokeAuthSession(id: string) {
  return request.delete<ApiResponse<{ success: boolean }>>(`/auth/sessions/${id}`)
}

export function logoutAllSessions() {
  return request.post<ApiResponse<{ success: boolean }>>('/auth/logout-all')
}

export function startPasswordRecovery(payload: PasswordRecoveryStartPayload) {
  return request.post<ApiResponse<PasswordRecoveryStartResponse>>('/auth/password-recovery/start', payload)
}

export function resetPasswordByRecovery(payload: PasswordRecoveryResetPayload) {
  return request.post<ApiResponse<PasswordRecoveryResetResponse>>('/auth/password-recovery/reset', payload)
}
