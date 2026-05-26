<template>
  <div class="page-card dashboard-page" v-loading="refreshing">
    <div class="page-header page-header-hero" :class="heroClass">
      <div>
        <div class="hero-kicker">{{ heroKicker }}</div>
        <h2>{{ heroTitle }}</h2>
        <p>{{ heroDesc }}</p>
      </div>
      <el-space>
        <el-button :loading="refreshing" @click="refresh">刷新</el-button>
        <el-button v-if="primaryAction" type="primary" @click="go(primaryAction.path)">{{ primaryAction.title }}</el-button>
      </el-space>
    </div>

    <el-row :gutter="16">
      <el-col v-for="item in dashboard.summaryCards || []" :key="item.label" :span="6">
        <el-card shadow="never" class="summary-card">
          <div class="summary-topline">{{ item.label }}</div>
          <div class="summary-value">{{ item.value }}</div>
          <div class="summary-desc">{{ item.desc }}</div>
        </el-card>
      </el-col>
    </el-row>

    <template v-if="isPlatformAdmin">
      <el-row :gutter="16">
        <el-col :span="14">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div>
                  <div class="card-title">高风险机构</div>
                  <div class="card-subtitle">优先关注续费与恢复动作</div>
                </div>
                <el-button link type="primary" @click="go('/platform/subscriptions')">进入续费跟进台</el-button>
              </div>
            </template>
            <el-table :data="dashboard.platformRisks || []" size="small" border>
              <el-table-column prop="organizationName" label="机构名称" min-width="160" />
              <el-table-column prop="followUpStatus" label="跟进状态" min-width="140" />
              <el-table-column prop="remainingDays" label="剩余天数" width="110" />
              <el-table-column label="账号配额" min-width="120">
                <template #default="scope">{{ scope.row.usedUsers }}/{{ scope.row.maxUsers }}</template>
              </el-table-column>
              <el-table-column label="校区配额" min-width="120">
                <template #default="scope">{{ scope.row.usedCampuses }}/{{ scope.row.maxCampuses }}</template>
              </el-table-column>
              <el-table-column label="风险级别" width="110">
                <template #default="scope">{{ formatHealthLevel(scope.row.healthLevel) }}</template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
        <el-col :span="10">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div>
                  <div class="card-title">平台指标</div>
                  <div class="card-subtitle">机构和资源规模概况</div>
                </div>
              </div>
            </template>
            <div class="metric-stack">
              <div v-for="item in dashboard.metricItems || []" :key="item.label" class="metric-item">
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <template v-else-if="isTeacher">
      <el-row :gutter="16">
        <el-col :span="14">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div>
                  <div class="card-title">本周课程</div>
                  <div class="card-subtitle">按当前账号可见范围展示</div>
                </div>
                <el-button link type="primary" @click="go('/my-schedules')">查看全部</el-button>
              </div>
            </template>
            <el-table :data="dashboard.schedules || []" size="small" border>
              <el-table-column prop="courseName" label="课程" min-width="120" />
              <el-table-column prop="classDate" label="日期" min-width="100" />
              <el-table-column prop="startTime" label="开始时间" min-width="120" />
              <el-table-column prop="status" label="状态" min-width="90" />
            </el-table>
          </el-card>
        </el-col>
        <el-col :span="10">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div>
                  <div class="card-title">最近上课记录</div>
                  <div class="card-subtitle">回看课堂执行情况</div>
                </div>
                <el-button link type="primary" @click="go('/my-lesson-records')">查看全部</el-button>
              </div>
            </template>
            <div class="record-list" v-if="(dashboard.lessonRecords || []).length">
              <div v-for="item in dashboard.lessonRecords || []" :key="item.id" class="record-item">
                <div class="record-item-left">
                  <strong>{{ item.lessonContent || `上课记录 #${item.id}` }}</strong>
                  <span>{{ formatAttendanceStatus(item.attendanceStatus) }}</span>
                </div>
                <div class="record-item-status">{{ item.recordedAt || '-' }}</div>
              </div>
            </div>
            <el-empty v-else description="暂无上课记录" />
          </el-card>
        </el-col>
      </el-row>
    </template>

    <template v-else-if="isOrgAdmin">
      <el-row :gutter="16">
        <el-col :span="14">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div>
                  <div class="card-title">异常校区分布</div>
                  <div class="card-subtitle">优先定位风险最集中的校区并进入处理</div>
                </div>
                <el-button link type="primary" @click="go('/platform/campuses')">进入校区管理</el-button>
              </div>
            </template>
            <el-table :data="dashboard.campusRisks || []" size="small" border>
              <el-table-column prop="campusName" label="校区" min-width="140" />
              <el-table-column prop="riskCount" label="异常总数" width="110" />
              <el-table-column prop="activeUsers" label="启用账号" width="110" />
              <el-table-column label="主要异常" min-width="220">
                <template #default="scope">
                  低课时 {{ scope.row.lowLessonCount }} / 欠费 {{ scope.row.arrearsCount }} / 调补课 {{ scope.row.rescheduleCount }}
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
        <el-col :span="10">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div>
                  <div class="card-title">组织治理提示</div>
                  <div class="card-subtitle">优先关注跨校区的资源、账号与订阅状态</div>
                </div>
              </div>
            </template>
            <div class="metric-stack">
              <div class="metric-item">
                <span>管理视角</span>
                <strong>机构级</strong>
              </div>
              <div class="metric-item">
                <span>主工作区</span>
                <strong>校区 / 用户 / 订阅</strong>
              </div>
              <div class="metric-item">
                <span>处理方式</span>
                <strong>优先通过校区视角兜底</strong>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <template v-else>
      <el-row :gutter="16">
        <el-col :span="14">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div>
                  <div class="card-title">最新收入动态</div>
                  <div class="card-subtitle">关注最近收费和续费节奏</div>
                </div>
                <el-button link type="primary" @click="go(paymentPath)">进入收费管理</el-button>
              </div>
            </template>
            <div class="payment-list" v-if="(dashboard.payments || []).length">
              <div v-for="item in dashboard.payments || []" :key="item.id" class="payment-item">
                <div class="payment-item-left">
                  <strong>{{ item.studentName || `缴费记录 #${item.id}` }}</strong>
                  <span>{{ [formatPaymentType(item.paymentType), item.lessonPackageName].filter(Boolean).join(' · ') }}</span>
                </div>
                <div class="payment-item-amount">¥ {{ Number(item.amount || 0).toFixed(2) }}</div>
              </div>
            </div>
            <el-empty v-else description="暂无缴费记录" />
          </el-card>
        </el-col>
        <el-col :span="10">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div>
                  <div class="card-title">风险预警</div>
                  <div class="card-subtitle">先看最值得处理的异常信号</div>
                </div>
              </div>
            </template>
            <div class="metric-stack">
              <button
                v-for="item in dashboard.alerts || []"
                :key="item.label"
                class="metric-item metric-button"
                type="button"
                @click="item.path ? go(item.path) : undefined"
              >
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <el-card shadow="never" class="content-card quick-entry-card" v-if="(dashboard.quickActions || []).length">
      <template #header>
        <div class="card-header">
          <div>
            <div class="card-title">快捷动作</div>
            <div class="card-subtitle">把最常做的经营动作放到首页</div>
          </div>
        </div>
      </template>
      <div class="quick-grid">
        <button v-for="item in dashboard.quickActions || []" :key="item.path" class="quick-entry" type="button" @click="go(item.path)">
          <span class="quick-entry-title">{{ item.title }}</span>
          <span class="quick-entry-desc">{{ item.desc }}</span>
        </button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchDashboardData, type DashboardResponse, type DashboardQuickAction } from '@/api/dashboard'
import { useAuthStore } from '@/stores/auth'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const router = useRouter()
const authStore = useAuthStore()
const refreshing = ref(false)
const dashboard = ref<DashboardResponse>({ role: '', summaryCards: [] })

const role = computed(() => authStore.currentUser?.roleCode || '')
const isTeacher = computed(() => role.value === 'teacher')
const isPlatformAdmin = computed(() => role.value?.startsWith('platform_'))
const isOrgAdmin = computed(() => role.value === 'org_admin')
const paymentPath = computed(() => (isOrgAdmin.value ? '/org/payments' : '/campus/payments'))

const heroKicker = computed(() => {
  if (isTeacher.value) return '工作台'
  if (isPlatformAdmin.value) return '平台工作台'
  return '经营驾驶舱'
})
const heroTitle = computed(() => {
  if (isTeacher.value) return '我的教学'
  if (isPlatformAdmin.value) return role.value === 'platform_admin' ? '平台经营概览' : '平台岗位工作台'
  return isOrgAdmin.value ? '机构经营看板' : '校区执行看板'
})
const heroDesc = computed(() => {
  if (isTeacher.value) return '查看本周课程、课堂记录和调课申请，保持个人授课节奏清晰可见。'
  if (isPlatformAdmin.value) return role.value === 'platform_admin' ? '从机构生命周期、续费风险和平台资源视角查看整体运营情况。' : '按当前平台岗位聚合可见指标与常用入口，避免越权入口干扰。'
  return isOrgAdmin.value ? '围绕校区、账号、订阅和机构级风险信号，帮助机构负责人做跨校区治理与经营判断。' : '聚焦今日执行、到课与现场风险，帮助校区快速处理日常运营。'
})
const heroClass = computed(() => {
  if (isTeacher.value) return 'teacher-hero'
  if (isPlatformAdmin.value) return 'platform-hero'
  return ''
})
const primaryAction = computed<DashboardQuickAction | null>(() => {
  const actions = dashboard.value.quickActions || []
  return actions[0] || null
})
const shouldBlockDashboardRequest = computed(() => authStore.isAuthenticated && authStore.mustChangePassword)

onMounted(() => {
  refresh()
})

async function refresh() {
  if (shouldBlockDashboardRequest.value) {
    dashboard.value = { role: role.value, summaryCards: [] }
    refreshing.value = false
    return
  }

  refreshing.value = true
  try {
    const response = await fetchDashboardData()
    dashboard.value = response.data?.data || { role: '', summaryCards: [] }
  } catch (error) {
    message.error(extractErrorMessage(error, '首页数据加载失败，请稍后重试'))
  } finally {
    refreshing.value = false
  }
}

function formatAttendanceStatus(status?: string) {
  if (status === 'present') return '正常出勤'
  if (status === 'late') return '迟到'
  if (status === 'absent') return '缺席'
  if (status === 'leave') return '请假'
  return status || '-'
}

function formatPaymentType(type?: string) {
  if (type === 'signup') return '新签'
  if (type === 'renewal') return '续费'
  if (type === 'tuition') return '学费'
  return type || '收费'
}

function formatHealthLevel(level?: string) {
  if (level === 'high') return '高风险'
  if (level === 'medium') return '关注'
  return '健康'
}

function go(path: string) {
  router.push(path)
}
</script>

<style scoped>
.dashboard-page {
  display: grid;
  gap: 20px;
}

.page-header-hero {
  padding: 28px 32px;
  border-radius: 24px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(16, 185, 129, 0.08));
  border: 1px solid rgba(37, 99, 235, 0.12);
}

.teacher-hero {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.12), rgba(59, 130, 246, 0.08));
  border-color: rgba(16, 185, 129, 0.12);
}

.platform-hero {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.12), rgba(236, 72, 153, 0.08));
  border-color: rgba(139, 92, 246, 0.12);
}

.hero-kicker {
  display: inline-flex;
  margin-bottom: 12px;
  padding: 7px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.06em;
  color: #1d4ed8;
  background: rgba(255, 255, 255, 0.72);
}

.teacher-hero .hero-kicker {
  color: #047857;
}

.platform-hero .hero-kicker {
  color: #7c3aed;
}

.summary-card {
  min-height: 160px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.summary-topline {
  color: #667085;
  margin-bottom: 14px;
  font-size: 13px;
}

.summary-value {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 14px;
  color: #111827;
}

.summary-desc {
  color: #606266;
  line-height: 1.6;
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.card-title {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
}

.card-subtitle {
  margin-top: 6px;
  font-size: 13px;
  color: #667085;
}

.payment-list,
.record-list,
.metric-stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.payment-item,
.record-item,
.metric-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-radius: 14px;
  background: linear-gradient(180deg, #f8fafc 0%, #f3f7ff 100%);
  border: 1px solid rgba(59, 130, 246, 0.08);
}

.metric-button {
  width: 100%;
  cursor: pointer;
  text-align: left;
}

.record-item {
  background: linear-gradient(180deg, #f0fdf4 0%, #ecfeff 100%);
  border-color: rgba(16, 185, 129, 0.08);
}

.payment-item-left,
.record-item-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
  color: #475467;
}

.payment-item-amount,
.metric-item strong {
  font-size: 18px;
  font-weight: 700;
  color: #111827;
}

.record-item-status {
  font-size: 13px;
  color: #059669;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 14px;
}

.quick-entry {
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  border-radius: 18px;
  padding: 18px;
  text-align: left;
  cursor: pointer;
}

.quick-entry-title {
  display: block;
  font-size: 15px;
  font-weight: 700;
  color: #111827;
}

.quick-entry-desc {
  display: block;
  margin-top: 8px;
  font-size: 13px;
  line-height: 1.6;
  color: #667085;
}
</style>
