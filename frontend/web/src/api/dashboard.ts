import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface DashboardSummaryCard {
  label: string
  value: string
  desc: string
}

export interface DashboardQuickAction {
  title: string
  desc: string
  path: string
}

export interface DashboardAlert {
  label: string
  value: string
  level: string
  path?: string
}

export interface DashboardMetricItem {
  label: string
  value: string
}

export interface DashboardPlatformRiskItem {
  organizationName: string
  followUpStatus: string
  remainingDays: number
  usedUsers: number
  maxUsers: number
  usedCampuses: number
  maxCampuses: number
  healthLevel: string
}

export interface DashboardCampusRiskItem {
  campusId: number
  campusName: string
  organizationId: number
  organizationName: string
  activeUsers: number
  riskCount: number
  lowLessonCount: number
  arrearsCount: number
  inactiveCount: number
  rescheduleCount: number
}

export interface DashboardPaymentItem {
  id: number
  studentName: string
  lessonPackageName: string
  paymentType: string
  amount: number
}

export interface DashboardScheduleItem {
  id: number
  courseName: string
  classDate: string
  startTime: string
  status: string
}

export interface DashboardLessonRecordItem {
  id: number
  lessonContent: string
  attendanceStatus: string
  recordedAt: string
}

export interface DashboardResponse {
  role: string
  summaryCards: DashboardSummaryCard[]
  alerts?: DashboardAlert[]
  quickActions?: DashboardQuickAction[]
  payments?: DashboardPaymentItem[]
  schedules?: DashboardScheduleItem[]
  lessonRecords?: DashboardLessonRecordItem[]
  platformRisks?: DashboardPlatformRiskItem[]
  campusRisks?: DashboardCampusRiskItem[]
  metricItems?: DashboardMetricItem[]
  subscriptionDays?: number
}

export function fetchDashboardData(params?: { organizationId?: number; campusId?: number }) {
  return request.get<ApiResponse<DashboardResponse>>('/statistics/dashboard', { params })
}
