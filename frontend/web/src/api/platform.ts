import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface PlatformOrganizationItem {
  id: number
  code: string
  name: string
  status: string
  subscriptionStatus?: string
  planCode?: string
  editionCode?: string
  endsAt?: string
  daysRemaining?: number
  maxCampuses?: number
  maxUsers?: number
  usedCampuses?: number
  usedUsers?: number
  featureFlags?: string[]
  healthLevel?: string
  followUpStatus?: string
  followUpOwner?: string
  followUpRemark?: string
}

export interface PlatformCampusItem {
  id: number
  organizationId: number
  code: string
  name: string
  status: string
  planCode?: string
  usedUsers?: number
  daysRemaining?: number
  featureFlags?: string[]
  healthLevel?: string
}

export interface PlatformOrganizationPayload {
  orgName: string
  status?: string
  subscriptionStatus?: string
  planCode?: string
  endsAt?: string
  maxCampuses?: number
  maxUsers?: number
  timezone?: string
  remark?: string
}

export interface PlatformOrganizationUpdatePayload {
  orgName: string
  status?: string
  subscriptionStatus?: string
  planCode?: string
  endsAt?: string
  maxCampuses?: number
  maxUsers?: number
  timezone?: string
  remark?: string
}

export interface PlatformCampusPayload {
  organizationId?: number
  campusName: string
  status?: string
  remark?: string
}

export interface PlatformCampusUpdatePayload {
  organizationId?: number
  campusName: string
  status?: string
  remark?: string
}

export interface OrganizationSettings {
  organizationId: number
  organizationName: string
  timezone: string
  brandName: string
  notificationEmail: string
  securityPolicy?: string[]
  remark: string
  featureFlags?: string[]
  planCode?: string
  subscriptionStatus?: string
}

export interface OrganizationSettingsPayload {
  timezone?: string
  brandName?: string
  notificationEmail?: string
  securityPolicy?: string[]
  remark?: string
}

export function fetchPlatformOrganizations() {
  return request.get<ApiResponse<PlatformOrganizationItem[]>>('/platform/organizations')
}

export function createPlatformOrganization(payload: PlatformOrganizationPayload) {
  return request.post<ApiResponse<{ id: number }>>('/platform/organizations', payload)
}

export function updatePlatformOrganization(id: number, payload: PlatformOrganizationUpdatePayload) {
  return request.put<ApiResponse<{ rows_affected: number }>>(`/platform/organizations/${id}`, payload)
}

export function fetchPlatformCampuses() {
  return request.get<ApiResponse<PlatformCampusItem[]>>('/platform/campuses')
}

export function createPlatformCampus(payload: PlatformCampusPayload) {
  return request.post<ApiResponse<{ id: number }>>('/platform/campuses', payload)
}

export function updatePlatformCampus(id: number, payload: PlatformCampusUpdatePayload) {
  return request.put<ApiResponse<{ rows_affected: number }>>(`/platform/campuses/${id}`, payload)
}

export function fetchOrganizationSettings() {
  return request.get<ApiResponse<OrganizationSettings>>('/organization/settings')
}

export function updateOrganizationSettings(payload: OrganizationSettingsPayload) {
  return request.put<ApiResponse<{ rows_affected: number }>>('/organization/settings', payload)
}
