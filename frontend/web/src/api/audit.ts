import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import { useAuthStore } from '@/stores/auth'

export interface AuditLogItem {
  time: string
  level: string
  message: string
  action: string
  module: string
  targetId: number
  actorId: number
  actorUsername: string
  actorRole: string
  traceId: string
  detail?: Record<string, unknown>
}

export interface AuditLogListResponse {
  items: AuditLogItem[]
  total: number
  offset: number
  limit: number
}

export interface AuditLogMetaResponse {
  actions: string[]
  modules: string[]
}

export interface AuditTargetOption {
  id: number
  module: string
  name: string
  label: string
  extra?: string
}

export interface AuditTargetSearchResponse {
  items: AuditTargetOption[]
}

export interface AuditLogQuery {
  action?: string
  module?: string
  actorId?: number
  actorUsername?: string
  targetId?: number
  traceId?: string
  startTime?: string
  endTime?: string
  offset?: number
  limit?: number
}

export function fetchAuditLogs(params?: AuditLogQuery) {
  return request.get<ApiResponse<AuditLogListResponse>>('/audit/logs', { params })
}

export function fetchAuditLogMeta() {
  return request.get<ApiResponse<AuditLogMetaResponse>>('/audit/logs/meta')
}

export function searchAuditTargets(params?: { module?: string; keyword?: string; limit?: number }) {
  return request.get<ApiResponse<AuditTargetSearchResponse>>('/audit/logs/targets', { params })
}

export async function exportAuditLogs(params?: AuditLogQuery) {
  const authStore = useAuthStore()
  const response = await request.get<Blob>('/audit/logs/export', {
    params,
    responseType: 'blob',
    headers: {
      Accept: 'text/csv',
      ...(authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}),
      ...(authStore.refreshToken ? { 'X-Refresh-Token': authStore.refreshToken } : {}),
    },
  })
  return response.data
}
