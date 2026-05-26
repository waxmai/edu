import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface RecoveryAlertItem {
  recordedAt: string
  challengeId: string
  challengeStatus: string
  userId: number
  username: string
  roleCode: string
  channel: string
  targetMasked: string
  error: string
  expiresAt: string
}

export interface RecoveryAlertQuery {
  username?: string
  channel?: string
  challengeId?: string
  targetMasked?: string
  offset?: number
  limit?: number
}

export interface RecoveryAlertListResponse {
  items: RecoveryAlertItem[]
  total: number
  offset: number
  limit: number
}

export interface RecoveryAlertSummaryResponse {
  total: number
  byChannel: Record<string, number>
  byErrorCategory: Record<string, number>
  trend: {
    last24Hours: number
    last7Days: number
  }
  topErrors: Array<{ label: string; count: number }>
  topUsernames: Array<{ label: string; count: number }>
  topChannels: Array<{ label: string; count: number }>
  daily: Array<{ date: string; count: number }>
}

export function fetchRecoveryAlerts(params?: RecoveryAlertQuery) {
  return request.get<ApiResponse<RecoveryAlertListResponse>>('/recovery-alerts', { params })
}

export function fetchRecoveryAlertSummary(params?: RecoveryAlertQuery) {
  return request.get<ApiResponse<RecoveryAlertSummaryResponse>>('/recovery-alerts/summary', { params })
}
