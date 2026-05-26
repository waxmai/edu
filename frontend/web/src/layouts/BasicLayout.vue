<template>
  <div class="layout-shell">
    <aside class="layout-sidebar">
      <div class="layout-logo-wrap">
        <div class="layout-logo-mark" aria-label="学枢云品牌标识">
          <svg viewBox="0 0 44 44" role="img" aria-hidden="true">
            <rect width="44" height="44" rx="14" fill="url(#brandLogoBg)" />
            <path d="M13 15.5c0-2.5 2-4.5 4.5-4.5h11c2.5 0 4.5 2 4.5 4.5v15c0 2.5-2 4.5-4.5 4.5h-11c-2.5 0-4.5-2-4.5-4.5v-15Z" fill="#eff6ff" />
            <path d="M19 19h9M19 23.5h7M19 28h8" stroke="#2563eb" stroke-width="2.4" stroke-linecap="round" />
            <path d="M30 12.2l6 3.1-6 3.1-6-3.1 6-3.1Z" fill="#facc15" />
            <path d="M25.5 17c1.7 1.6 7.3 1.6 9 0v3.3c-1.9 1.5-7.1 1.5-9 0V17Z" fill="#f59e0b" />
            <circle cx="32.5" cy="29" r="5.8" fill="#10b981" />
            <path d="M29.6 29l1.8 1.8 4-4.1" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            <defs>
              <linearGradient id="brandLogoBg" x1="5" y1="4" x2="39" y2="40" gradientUnits="userSpaceOnUse">
                <stop stop-color="#38bdf8" />
                <stop offset="0.55" stop-color="#2563eb" />
                <stop offset="1" stop-color="#1d4ed8" />
              </linearGradient>
            </defs>
          </svg>
        </div>
        <div class="layout-logo-copy">
          <div class="layout-logo">学枢云</div>
          <div class="layout-subtitle">教培排课、课时与校区运营中枢</div>
        </div>
      </div>
      <nav class="layout-nav">
        <div v-for="section in navSections" :key="section.title" class="layout-nav-section">
          <div v-if="section.title" class="layout-nav-section-title">{{ section.title }}</div>
          <RouterLink v-for="item in section.items" :key="item.to" :to="item.to">{{ item.label }}</RouterLink>
        </div>
      </nav>
      <div class="layout-sidebar-footer">
        <div class="sidebar-role">当前身份</div>
        <div class="sidebar-role-value">{{ authStore.currentUser?.roleCode || 'guest' }}</div>
        <div v-if="viewAsBanner" class="sidebar-view-as-banner">
          <div class="sidebar-view-as-label">当前视角</div>
          <div class="sidebar-view-as-value">{{ viewAsBanner }}</div>
          <button class="sidebar-view-as-exit" type="button" @click="clearPlatformViewAs">退出视角</button>
        </div>
        <div class="sidebar-scope">数据范围：{{ authStore.currentUser?.dataScope || '-' }}</div>
        <div class="sidebar-scope">机构：{{ authStore.currentUser?.organizationName || '-' }}</div>
        <div class="sidebar-scope">校区：{{ authStore.currentUser?.campusName || '-' }}</div>
      </div>
    </aside>
    <main class="layout-main">
      <header class="layout-header">
        <div>
          <div class="layout-header-title">{{ pageTitle }}</div>
          <div class="layout-header-subtitle">{{ headerSubtitle }}</div>
          <div v-if="tenantBanner" class="layout-tenant-banner" :class="tenantBanner.type">{{ tenantBanner.text }}</div>
        </div>
        <div class="header-actions">
          <el-dropdown trigger="click" popper-class="app-user-dropdown" @command="handleUserMenuCommand">
            <button class="user-menu-trigger" type="button">
              <span class="user-menu-avatar">{{ userInitial }}</span>
              <span class="user-menu-meta">
                <span class="user-menu-name">{{ authStore.currentUser?.realName || authStore.currentUser?.username || '未登录' }}</span>
                <span class="user-menu-role">{{ userRoleLabel }}</span>
              </span>
              <span class="user-menu-arrow">▾</span>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <div class="user-dropdown-panel">
                  <div class="user-dropdown-summary">
                    <div class="user-dropdown-summary__name">{{ authStore.currentUser?.realName || authStore.currentUser?.username || '未登录' }}</div>
                    <div class="user-dropdown-summary__meta">
                      <span>{{ userRoleLabel }}</span>
                      <span v-if="authStore.currentUser?.username">@{{ authStore.currentUser?.username }}</span>
                    </div>
                  </div>
                  <el-dropdown-item command="change-password">修改密码</el-dropdown-item>
                  <el-dropdown-item command="logout" divided class="user-dropdown-danger">退出登录</el-dropdown-item>
                </div>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>
      <section class="layout-content">
        <RouterView />
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { filterMenuItems } from '@/composables/usePermission'
import { openPasswordDialog } from '@/stores/password-dialog'
import { confirmDialog } from '@/utils/confirm'
import { extractErrorMessage, isCancelError } from '@/utils/error'
import { message } from '@/utils/message'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

if (authStore.currentUser?.subscriptionStatus === 'expired') {
  message.warning('当前机构订阅已过期，关键写操作已被限制')
} else if (authStore.currentUser?.subscriptionStatus === 'past_due') {
  message.warning('当前机构订阅待续费，部分操作可能受限')
} else if (authStore.currentUser?.subscriptionStatus === 'suspended') {
  message.error('当前机构已停用，请联系平台管理员')
}

interface NavItem {
  to: string
  label: string
  roles?: string[]
  permission?: string
  menuPermission?: string
  featureFlag?: string
}

interface NavSection {
  title: string
  items: NavItem[]
}

function visibleSection(title: string, items: NavItem[]): NavSection | null {
  const visibleItems = filterMenuItems(items)
  return visibleItems.length ? { title, items: visibleItems } : null
}

function compactSections(sections: Array<NavSection | null>) {
  return sections.filter((section): section is NavSection => Boolean(section))
}

const navSections = computed(() => {
  const role = authStore.currentUser?.roleCode || ''
  const isOrgAdmin = role === 'org_admin'
  const isCampusAdmin = role === 'campus_admin'
  const isTeacher = role === 'teacher'

  if (isTeacher) {
    return compactSections([
      visibleSection('教学工作台', [{ to: '/dashboard', label: '工作台', menuPermission: 'menu:dashboard:view' }]),
      visibleSection('我的教学', [
        { to: '/my-students', label: '我的学生', permission: 'student:list', menuPermission: 'menu:student:view' },
        { to: '/my-schedules', label: '我的排课', permission: 'schedule:list', menuPermission: 'menu:schedule:view' },
        { to: '/my-lesson-records', label: '上课记录', permission: 'lesson-record:list', menuPermission: 'menu:lesson-record:view' },
        { to: '/my-reschedules', label: '调课申请', permission: 'reschedule:list', menuPermission: 'menu:reschedule:view' },
      ]),
    ])
  }

  if (role.startsWith('platform_')) {
    return compactSections([
      visibleSection('平台工作台', [{ to: '/dashboard', label: '平台概览', menuPermission: 'menu:dashboard:view' }]),
      visibleSection('平台运营', [
        { to: '/platform/ops', label: '运营总览', roles: ['platform_admin', 'platform_ops'], permission: 'platform:ops:view', menuPermission: 'menu:platform:view' },
        { to: '/platform/subscriptions', label: '续费跟进', roles: ['platform_admin', 'platform_ops', 'platform_finance'], permission: 'platform:subscription:list', menuPermission: 'menu:platform:view', featureFlag: 'subscription_center' },
        { to: '/platform/tenants', label: '机构管理', roles: ['platform_admin', 'platform_ops', 'platform_finance', 'platform_support'], permission: 'platform:tenant:list', menuPermission: 'menu:platform:orgs:view' },
        { to: '/platform/campuses', label: '校区管理', roles: ['platform_admin', 'platform_ops', 'platform_support'], permission: 'platform:campus:list', menuPermission: 'menu:campus:manage:view' },
      ]),
      visibleSection('账号与安全', [
        { to: '/platform/users', label: '平台用户', roles: ['platform_admin'], permission: 'system:user:list', menuPermission: 'menu:system:user:view' },
        { to: '/platform/audit-logs', label: '审计日志', roles: ['platform_admin', 'platform_auditor'], permission: 'platform:audit:list', menuPermission: 'menu:platform:view', featureFlag: 'audit_export' },
        { to: '/platform/recovery-alerts', label: '恢复失败告警', roles: ['platform_admin', 'platform_auditor', 'platform_support'], permission: 'platform:recovery-alert:list', menuPermission: 'menu:platform:view', featureFlag: 'recovery_ops' },
      ]),
    ])
  }

  if (isOrgAdmin) {
    return compactSections([
      visibleSection('机构工作台', [{ to: '/dashboard', label: '首页', menuPermission: 'menu:dashboard:view' }]),
      visibleSection('机构治理', [
        { to: '/platform/campuses', label: '校区管理', permission: 'platform:campus:list', menuPermission: 'menu:campus:manage:view' },
        { to: '/org/users', label: '用户管理', permission: 'system:user:list', menuPermission: 'menu:system:user:view', featureFlag: 'user_management' },
        { to: '/org/settings', label: '机构设置', permission: 'tenant:settings:view', menuPermission: 'menu:dashboard:view' },
        { to: '/org/subscription', label: '订阅中心', roles: ['org_admin'], permission: 'auth:me:view', menuPermission: 'menu:subscription:center:view' },
      ]),
      visibleSection('教务运营', [
        { to: '/students', label: '学员管理', permission: 'student:list', menuPermission: 'menu:student:view' },
        { to: '/courses', label: '课程定义', permission: 'course:list', menuPermission: 'menu:course:view' },
        { to: '/org/lesson-packages', label: '课时包管理', permission: 'lesson-package:list', menuPermission: 'menu:lesson-package:view' },
        { to: '/org/schedules', label: '排课管理', permission: 'schedule:list', menuPermission: 'menu:schedule:view' },
        { to: '/org/lesson-records', label: '上课记录', permission: 'lesson-record:list', menuPermission: 'menu:lesson-record:view' },
        { to: '/org/reschedules', label: '调补课管理', permission: 'reschedule:list', menuPermission: 'menu:reschedule:view' },
      ]),
      visibleSection('财务运营', [{ to: '/org/payments', label: '收费管理', permission: 'payment:list', menuPermission: 'menu:payment:view' }]),
    ])
  }

  return compactSections([
    visibleSection('校区工作台', [{ to: '/dashboard', label: '首页', menuPermission: 'menu:dashboard:view' }]),
    visibleSection('教务运营', [
      { to: '/students', label: '学员管理', permission: 'student:list', menuPermission: 'menu:student:view' },
      { to: '/courses', label: '课程定义', permission: 'course:list', menuPermission: 'menu:course:view' },
      { to: isCampusAdmin ? '/campus/lesson-packages' : '/org/lesson-packages', label: '课时包管理', permission: 'lesson-package:list', menuPermission: 'menu:lesson-package:view' },
      { to: isCampusAdmin ? '/campus/schedules' : '/org/schedules', label: '排课管理', permission: 'schedule:list', menuPermission: 'menu:schedule:view' },
      { to: isCampusAdmin ? '/campus/lesson-records' : '/org/lesson-records', label: '上课记录', permission: 'lesson-record:list', menuPermission: 'menu:lesson-record:view' },
      { to: isCampusAdmin ? '/campus/reschedules' : '/org/reschedules', label: '调补课管理', permission: 'reschedule:list', menuPermission: 'menu:reschedule:view' },
    ]),
    visibleSection('校区经营', [{ to: isCampusAdmin ? '/campus/payments' : '/org/payments', label: '收费管理', permission: 'payment:list', menuPermission: 'menu:payment:view' }]),
    visibleSection('账号管理', [{ to: isCampusAdmin ? '/campus/users' : '/org/users', label: '用户管理', permission: 'system:user:list', menuPermission: 'menu:system:user:view', featureFlag: 'user_management' }]),
  ])
})

const pageTitle = computed(() => String(route.meta.title || '学枢云'))
const headerSubtitle = computed(() => {
  const role = authStore.currentUser?.roleCode || ''
  const campusName = authStore.currentUser?.campusName
  const orgName = authStore.currentUser?.organizationName
  if (role === 'platform_admin') return '平台治理与全局监控'
  if (role === 'org_admin') return `${orgName || '机构'}运营管理`
  if (role === 'campus_admin') return `${campusName || '校区'}运营管理`
  if (role === 'teacher') return '个人教学工作台'
  return '统一管理教培业务流程、账号安全与关键数据'
})
const userRoleLabel = computed(() => {
  if (authStore.currentUser?.roleCode === 'platform_admin') return '平台超管'
  if (authStore.currentUser?.roleCode === 'org_admin') return '机构超管'
  if (authStore.currentUser?.roleCode === 'campus_admin') return '校区管理员'
  if (authStore.currentUser?.roleCode === 'teacher') return '教师账号'
  return '当前账号'
})
const tenantBanner = computed(() => {
  const status = authStore.currentUser?.subscriptionStatus || ''
  if (status === 'expired') {
    return { type: 'danger', text: '当前机构订阅已过期，关键写操作已被限制，请尽快续费。' }
  }
  if (status === 'past_due') {
    return { type: 'warning', text: '当前机构订阅待续费，关键写操作可能受限。' }
  }
  if (status === 'suspended') {
    return { type: 'danger', text: '当前机构已停用，请联系平台管理员恢复。' }
  }
  if (status === 'trial') {
    return { type: 'info', text: '当前机构处于试用期，请留意套餐与有效期。' }
  }
  return null
})
const userInitial = computed(() => {
  const source = authStore.currentUser?.realName || authStore.currentUser?.username || 'U'
  return source.trim().slice(0, 1).toUpperCase()
})
const viewAsBanner = computed(() => {
  if (authStore.currentUser?.roleCode !== 'platform_admin') return ''
  const current = authStore.platformViewAs
  if (!current) return ''
  const org = current.organizationName || (current.organizationId ? `机构 #${current.organizationId}` : '')
  const campus = current.campusName || (current.campusId ? `校区 #${current.campusId}` : '')
  return [org, campus].filter(Boolean).join(' / ')
})

function clearPlatformViewAs() {
  authStore.setPlatformViewAs(null)
  message.success('已退出机构视角')
}

async function handleLogout() {
  try {
    await confirmDialog('确认退出当前登录状态吗？', '退出登录确认', {
      type: 'warning',
      confirmButtonText: '确认退出',
      cancelButtonText: '取消',
    })
    await authStore.logout()
    message.success('已退出登录')
    await router.push('/login')
  } catch (error) {
    if (isCancelError(error)) {
      return
    }
    message.error(extractErrorMessage(error, '退出登录失败，请稍后重试'))
  }
}

async function handleUserMenuCommand(command: string) {
  if (command === 'change-password') {
    openPasswordDialog(false)
    return
  }
  if (command === 'logout') {
    await handleLogout()
  }
}
</script>

<style scoped>
.layout-shell {
  display: flex;
  min-height: 100vh;
}

.layout-sidebar {
  width: 252px;
  background: linear-gradient(180deg, #0f172a 0%, #162033 100%);
  color: #fff;
  padding: 22px 16px 18px;
  box-shadow: inset -1px 0 0 rgba(255, 255, 255, 0.04);
  display: flex;
  flex-direction: column;
}

.layout-logo-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
  padding: 12px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.06);
}

.layout-logo-mark {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.24);
  overflow: hidden;
  flex: 0 0 auto;
}

.layout-logo-mark svg {
  display: block;
  width: 44px;
  height: 44px;
}

.layout-logo-copy {
  min-width: 0;
}

.layout-logo {
  font-size: 18px;
  font-weight: 700;
  line-height: 1.2;
}

.layout-subtitle {
  margin-top: 4px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.62);
}

.layout-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.layout-nav-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.layout-nav-section + .layout-nav-section {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.layout-nav-section-title {
  padding: 0 12px 2px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: rgba(203, 213, 225, 0.62);
}

.layout-nav a {
  color: rgba(255, 255, 255, 0.82);
  padding: 11px 12px;
  border-radius: 12px;
  transition: background 0.18s ease, color 0.18s ease, transform 0.18s ease;
}

.layout-nav a:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
  transform: translateX(2px);
}

.layout-nav a.router-link-active {
  color: #ffffff;
  font-weight: 600;
  background: linear-gradient(90deg, rgba(59, 130, 246, 0.34), rgba(59, 130, 246, 0.12));
  box-shadow: inset 0 0 0 1px rgba(96, 165, 250, 0.15);
}

.layout-sidebar-footer {
  margin-top: auto;
  padding: 14px 12px 4px;
  color: rgba(255, 255, 255, 0.66);
}

.sidebar-role {
  font-size: 12px;
}

.sidebar-role-value {
  margin-top: 6px;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
}

.sidebar-view-as-banner {
  margin-top: 10px;
  padding: 10px;
  border-radius: 12px;
  background: rgba(59, 130, 246, 0.16);
  border: 1px solid rgba(96, 165, 250, 0.2);
}

.sidebar-view-as-label {
  font-size: 11px;
  color: rgba(191, 219, 254, 0.9);
}

.sidebar-view-as-value {
  margin-top: 4px;
  font-size: 12px;
  color: #eff6ff;
  line-height: 1.5;
}

.sidebar-view-as-exit {
  margin-top: 8px;
  border: 0;
  background: transparent;
  color: #bfdbfe;
  font-size: 12px;
  padding: 0;
  cursor: pointer;
}

.sidebar-scope {
  margin-top: 6px;
  font-size: 12px;
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.layout-header {
  min-height: 80px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  background: rgba(255, 255, 255, 0.78);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid rgba(15, 23, 42, 0.06);
  color: #111827;
}

.layout-header-title {
  font-size: 22px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.layout-header-subtitle {
  margin-top: 6px;
  font-size: 13px;
  color: #6b7280;
}

.layout-tenant-banner {
  margin-top: 8px;
  font-size: 12px;
  font-weight: 500;
}

.layout-tenant-banner.info {
  color: #2563eb;
}

.layout-tenant-banner.warning {
  color: #b45309;
}

.layout-tenant-banner.danger {
  color: #dc2626;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  font-weight: 400;
}

.user-menu-trigger {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-width: 220px;
  padding: 8px 14px 8px 10px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.96));
  border-radius: 18px;
  cursor: pointer;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
  transition: all 0.18s ease;
}

.user-menu-trigger:hover {
  border-color: rgba(59, 130, 246, 0.24);
  box-shadow: 0 16px 30px rgba(37, 99, 235, 0.12);
  transform: translateY(-1px);
}

.user-menu-avatar {
  width: 38px;
  height: 38px;
  border-radius: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.22);
}

.user-menu-meta {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
}

.user-menu-name {
  width: 100%;
  color: #111827;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.3;
  text-align: left;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-menu-role {
  margin-top: 3px;
  color: #667085;
  font-size: 12px;
  line-height: 1.2;
}

.user-menu-arrow {
  color: #98a2b3;
  font-size: 12px;
  flex: 0 0 auto;
}

.layout-content {
  padding: 28px;
  width: 100%;
  max-width: 1680px;
}
</style>
