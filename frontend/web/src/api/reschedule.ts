import request from '@/utils/request'
import type { ApiListResult, ApiResponse } from '@/types/api'

export interface RescheduleRecordItem {
  id: number
  oldScheduleId: number
  newScheduleId?: number
  operationType: string
  reason?: string
  operatorId?: number
  createdAt?: string
}

export interface RescheduleQuery {
  studentId?: number | string
  operationType?: string
  startDate?: string
  endDate?: string
  pageNum?: number
  pageSize?: number
}

export function fetchRescheduleList(params?: RescheduleQuery) {
  return request.get<ApiResponse<ApiListResult<RescheduleRecordItem> | RescheduleRecordItem[]>>('/reschedule-records', { params })
}
