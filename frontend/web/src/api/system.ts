import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface SystemUserItem {
  id: number
  username: string
  roleCode: string
  organizationId?: number
  campusId?: number
  realName: string
  phone?: string
  email?: string
  status: string
  failedLoginCount?: number
  lockedUntil?: string
  mustChangePassword?: boolean
  remark?: string
  createdAt?: string
  updatedAt?: string
}

export interface SystemUserListPayload {
  list: SystemUserItem[]
  total: number
}

export interface SystemUserQuery {
  username?: string
  realName?: string
  roleCode?: string
  status?: string
  organizationId?: number
  campusId?: number
  pageNum?: number
  pageSize?: number
}

export interface SystemUserFormPayload {
  username: string
  password?: string
  roleCode: string
  organizationId?: number
  campusId?: number
  realName: string
  phone?: string
  status: string
  remark?: string
}

export interface ResetPasswordPayload {
  newPassword?: string
}

export function fetchSystemUsers(params?: SystemUserQuery) {
  return request.get<ApiResponse<SystemUserListPayload>>('/users', { params })
}

export function createSystemUser(payload: SystemUserFormPayload) {
  return request.post<ApiResponse<{ id: number }>>('/users', payload)
}

export function updateSystemUser(id: number, payload: SystemUserFormPayload) {
  return request.put<ApiResponse<{ rows_affected: number }>>(`/users/${id}`, payload)
}

export function updateSystemUserStatus(id: number, status: string) {
  return request.patch<ApiResponse<{ rows_affected: number }>>(`/users/${id}/status`, { status })
}

export function resetSystemUserPassword(id: number, payload?: ResetPasswordPayload) {
  return request.post<ApiResponse<{ rows_affected: number }>>(`/users/${id}/reset-password`, payload || {})
}

export function unlockSystemUser(id: number) {
  return request.post<ApiResponse<{ rows_affected: number }>>(`/users/${id}/unlock`)
}
