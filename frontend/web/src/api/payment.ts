import request from '@/utils/request'
import type { PaymentFormModel, PaymentItem, PaymentSearchParams } from '@/modules/payment/types/payment'
import type { ApiListResult, ApiResponse } from '@/types/api'

export function fetchPaymentList(params?: PaymentSearchParams) {
  return request.get<ApiResponse<ApiListResult<PaymentItem> | PaymentItem[]>>('/payment-records', { params })
}

export function createPayment(payload: PaymentFormModel) {
  return request.post<ApiResponse<{ id: number }>>('/payment-records', payload)
}

export function updatePayment(id: number | string, payload: Partial<PaymentFormModel>) {
  return request.put<ApiResponse<{ rows_affected: number }>>(`/payment-records/${id}`, payload)
}

export function deletePayment(id: number | string) {
  return request.delete<ApiResponse<{ rows_affected: number }>>(`/payment-records/${id}`)
}

export function fetchStudentPaymentHistory(studentId: number | string) {
  return request.get<ApiResponse<PaymentItem[]>>(`/students/${studentId}/payment-records`)
}
