import request from '@/utils/request'
import type { ApiListResult, ApiResponse } from '@/types/api'

export interface CourseItem {
  id: number
  courseName: string
  subject: string
  courseType: string
  durationMinutes: number
  feeStandard: number
  status: string
  remark?: string
  createdAt?: string
  updatedAt?: string
}

export interface CourseFormModel {
  courseName: string
  subject: string
  courseType: string
  durationMinutes: number
  feeStandard: number
  status: string
  remark: string
}

export interface CourseSearchParams {
  courseName?: string
  subject?: string
  status?: string
  pageNum?: number
  pageSize?: number
}

export function fetchCourseList(params?: CourseSearchParams) {
  return request.get<ApiResponse<ApiListResult<CourseItem> | CourseItem[]>>('/courses', { params })
}

export function createCourse(payload: CourseFormModel) {
  return request.post<ApiResponse<{ id: number }>>('/courses', payload)
}

export function updateCourse(id: number | string, payload: CourseFormModel) {
  return request.put<ApiResponse<{ rows_affected: number }>>(`/courses/${id}`, payload)
}

export function deleteCourse(id: number | string) {
  return request.delete<ApiResponse<{ rows_affected: number }>>(`/courses/${id}`)
}
