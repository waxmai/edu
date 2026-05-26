import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface SubscriptionFeatureEntitlement {
  code: string
  name: string
  enabled: boolean
  description: string
}

export interface SubscriptionOverview {
  organizationId?: number
  organizationName?: string
  planCode: string
  planName: string
  status: string
  startsAt?: string
  endsAt?: string
  remainingDays: number
  maxUsers: number
  usedUsers: number
  maxCampuses: number
  usedCampuses: number
  featureFlags?: string[]
  featureEntitlements?: SubscriptionFeatureEntitlement[]
  editionName: string
  renewalHint: string
  upgradeHint: string
}

export interface SubscriptionPipelineItem {
  organizationId: number
  organizationName: string
  subscriptionStatus: string
  planCode: string
  planName: string
  endsAt?: string
  remainingDays: number
  usedUsers: number
  maxUsers: number
  usedCampuses: number
  maxCampuses: number
  healthLevel: string
  followUpStatus: string
  followUpOwner: string
  followUpNote: string
  lastContactAt?: string
  followUpRemark: string
  featureFlags?: string[]
  editionName: string
}

export interface SubscriptionFollowUpPayload {
  followUpStatus: string
  followUpOwner?: string
  followUpNote?: string
  lastContactAt?: string
}

export function fetchCurrentSubscription() {
  return request.get<ApiResponse<SubscriptionOverview>>('/subscriptions/current')
}

export function fetchSubscriptionPipeline() {
  return request.get<ApiResponse<SubscriptionPipelineItem[]>>('/platform/subscriptions')
}

export function updateSubscriptionFollowUp(organizationId: number, payload: SubscriptionFollowUpPayload) {
  return request.put<ApiResponse<{ rows_affected: number }>>(`/platform/subscriptions/${organizationId}/follow-up`, payload)
}
