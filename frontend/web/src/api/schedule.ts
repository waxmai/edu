import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { ScheduleItem } from '@/modules/schedule/types/schedule'

export interface ScheduleListItem extends ScheduleItem {
  studentId?: number
  courseId?: number
  teacherId?: number
  lessonPackageId?: number
  classDate?: string
  classroom?: string
  scheduleStatus?: string
  remark?: string
}

export interface ScheduleCreatePayload {
  studentId: number
  courseId: number
  teacherId: number
  lessonPackageId?: number | null
  classDate: string
  startTime: string
  endTime: string
  classroom?: string
  scheduleStatus?: string
  remark?: string
}

export interface MakeupScheduleCreatePayload {
  originalScheduleId: number
  studentId: number
  teacherId: number
  courseId: number
  lessonPackageId?: number | null
  classDate: string
  startTime: string
  endTime: string
  classroom?: string
  remark?: string
}

export interface ScheduleReschedulePayload {
  reason: string
  newClassDate: string
  newStartTime: string
  newEndTime: string
  classroom?: string
}

export interface ScheduleCancelPayload {
  reason: string
}

export interface ScheduleLeavePayload {
  reason: string
}

export function fetchScheduleList(params?: Record<string, unknown>) {
  return request.get<ApiResponse<ScheduleListItem[]>>('/schedules', { params })
}

export function createSchedule(payload: ScheduleCreatePayload) {
  return request.post<ApiResponse<{ id: number }>>('/schedules', payload)
}

export function deleteSchedule(id: number | string) {
  return request.delete<ApiResponse<{ rows_affected: number }>>(`/schedules/${id}`)
}

export function cancelSchedule(id: number | string, payload: ScheduleCancelPayload) {
  return request.patch<ApiResponse<{ rows_affected: number }>>(`/schedules/${id}/cancel`, payload)
}

export function leaveSchedule(id: number | string, payload: ScheduleLeavePayload) {
  return request.post<ApiResponse<{ id: number }>>(`/schedules/${id}/leave`, payload)
}

export function rescheduleSchedule(id: number | string, payload: ScheduleReschedulePayload) {
  return request.post<ApiResponse<{ id: number }>>(`/schedules/${id}/reschedule`, payload)
}

export function createMakeupSchedule(payload: MakeupScheduleCreatePayload) {
  return request.post<ApiResponse<{ id: number }>>('/makeup-schedules', payload)
}
