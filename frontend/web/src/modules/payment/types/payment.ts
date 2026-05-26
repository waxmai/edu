export interface PaymentSearchParams {
  studentId?: number | string
  lessonPackageId?: number | string
  paymentType?: string
  paymentStatus?: string
  paymentMethod?: string
  startDate?: string
  endDate?: string
  pageNum?: number
  pageSize?: number
}

export interface PaymentItem {
  id: number
  studentId: number
  studentName?: string
  lessonPackageId: number
  lessonPackageName?: string
  courseName?: string
  paymentType: string
  amount: number
  paymentMethod: string
  paymentTime: string
  paymentStatus: string
  remark?: string
}

export interface PaymentFormModel {
  studentId: number | null
  lessonPackageId: number | null
  paymentType: string
  amount: number
  paymentMethod: string
  paymentTime: string
  paymentStatus: string
  remark: string
}
