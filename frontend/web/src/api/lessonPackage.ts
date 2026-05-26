import request from '@/utils/request'
import type { LessonPackageFormModel, LessonPackageItem, LessonPackageSearchParams } from '@/modules/lesson-package/types/lesson-package'
import type { ApiListResult, ApiResponse } from '@/types/api'

export function fetchLessonPackageList(params?: LessonPackageSearchParams) {
  return request.get<ApiResponse<ApiListResult<LessonPackageItem> | LessonPackageItem[]>>('/lesson-packages', { params })
}

export function createLessonPackage(payload: LessonPackageFormModel) {
  return request.post<ApiResponse<{ id: number }>>('/lesson-packages', payload)
}

export type LessonPackageUpdatePayload = Omit<LessonPackageFormModel, 'studentId'>

export function updateLessonPackage(id: number | string, payload: LessonPackageUpdatePayload) {
  return request.put<ApiResponse<unknown>>(`/lesson-packages/${id}`, payload)
}

export function deleteLessonPackage(id: number | string) {
  return request.delete<ApiResponse<{ rows_affected: number }>>(`/lesson-packages/${id}`)
}
