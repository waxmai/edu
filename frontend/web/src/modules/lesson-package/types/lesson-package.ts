export interface LessonPackageSearchParams {
  studentId?: number | string
  courseId?: number | string
  status?: string
  lowLessonAlert?: boolean
  paymentStatus?: string
  pageNum?: number
  pageSize?: number
}

export interface LessonPackageItem {
  id: number
  studentId: number
  studentName?: string
  courseId: number
  courseName?: string
  subject?: string
  totalLessons: number
  usedLessons: number
  remainLessons: number
  totalAmount: number
  paidAmount: number
  startDate?: string
  endDate?: string
  status: string
  lowLessonThreshold?: number
  remark?: string
}

export interface LessonPackageFormModel {
  studentId: number | null
  courseId: number | null
  totalLessons: number
  totalAmount: number
  paidAmount: number
  startDate: string
  endDate: string
  status: string
  lowLessonThreshold: number
  remark: string
}
