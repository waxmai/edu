export interface StudentSearchParams {
  studentName?: string
  subject?: string
  status?: string
  parentPhone?: string
  inactiveAlert?: boolean
  pageNum?: number
  pageSize?: number
}

export interface StudentItem {
  id: number
  studentName: string
  gender?: string
  grade: string
  subject: string
  teachingType: string
  parentName: string
  parentPhone: string
  phone?: string
  status: string
  remark?: string
}

export interface StudentDetail extends StudentItem {
  createdAt?: string
}

export interface StudentFormModel {
  studentName: string
  gender: string
  grade: string
  phone: string
  parentName: string
  parentPhone: string
  subject: string
  teachingType: string
  status: string
  remark: string
}
