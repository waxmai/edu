import request from '@/utils/request'
import type { StudentDetail, StudentFormModel, StudentSearchParams } from '@/modules/student/types/student'
import type { ApiListResult, ApiResponse } from '@/types/api'

export function fetchStudentList(params?: StudentSearchParams) {
  return request.get<ApiResponse<ApiListResult<StudentDetail> | StudentDetail[]>>('/students', { params })
}

export function fetchStudentDetail(id: number | string) {
  return request.get<ApiResponse<StudentDetail>>(`/student/${id}`)
}

export function createStudent(payload: StudentFormModel) {
  return request.post<ApiResponse<{ id: number }>>('/student', payload)
}

export function updateStudent(id: number | string, payload: StudentFormModel) {
  return request.put<ApiResponse<unknown>>(`/student/${id}`, payload)
}

export function deleteStudent(id: number | string) {
  return request.delete<ApiResponse<{ rows_affected: number }>>(`/students/${id}`)
}
