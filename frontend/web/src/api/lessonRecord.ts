import request from '@/utils/request'
import type { ApiListResult, ApiResponse } from '@/types/api'

export interface LessonRecordItem {
  id: number
  scheduleId: number
  studentId: number
  teacherId: number
  attendanceStatus: string
  lessonContent?: string
  homework?: string
  feedback?: string
  lessonDeducted: boolean
  deductLessonCount: number
  recordedAt?: string
  updatedAt?: string
}

export interface LessonRecordFormModel {
  scheduleId: number | null
  studentId: number | null
  teacherId: number | null
  attendanceStatus: string
  lessonContent: string
  homework: string
  feedback: string
  needDeductLesson: boolean
  deductLessonCount: number
}

export interface LessonRecordSearchParams {
  studentId?: number | string
  teacherId?: number | string
  attendanceStatus?: string
  startDate?: string
  endDate?: string
  pageNum?: number
  pageSize?: number
}

export function fetchLessonRecordList(params?: LessonRecordSearchParams) {
  return request.get<ApiResponse<ApiListResult<LessonRecordItem> | LessonRecordItem[]>>('/lesson-records', { params })
}

export function createLessonRecord(payload: LessonRecordFormModel) {
  return request.post<ApiResponse<{ id: number }>>('/lesson-records', payload)
}

export function updateLessonRecord(id: number | string, payload: Partial<LessonRecordFormModel>) {
  return request.put<ApiResponse<{ rows_affected: number }>>(`/lesson-records/${id}`, payload)
}

export function deleteLessonRecord(id: number | string) {
  return request.delete<ApiResponse<{ rows_affected: number }>>(`/lesson-records/${id}`)
}
