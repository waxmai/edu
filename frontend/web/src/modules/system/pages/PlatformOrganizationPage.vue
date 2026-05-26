<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">SaaS 平台治理</div>
        <h2>机构管理</h2>
        <p>统一维护机构开通状态、套餐配置、资源配额与续费生命周期。</p>
      </div>
      <div class="actions">
        <el-button @click="loadOrganizations">刷新</el-button>
        <el-button v-if="canCreateTenant" type="primary" @click="openOrgCreate">新增机构</el-button>
      </div>
    </div>

    <el-alert
      title="平台管理员可在此维护机构基础信息、套餐与资源配额，并可直接进入机构视角排障。"
      type="warning"
      :closable="false"
    />

    <el-card shadow="never" class="section-card filter-card">
      <el-form inline>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="机构名称 / 编号" clearable />
        </el-form-item>
        <el-form-item label="机构状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="订阅状态">
          <el-select v-model="filters.subscriptionStatus" placeholder="全部" clearable style="width: 160px">
            <el-option label="试用中" value="trial" />
            <el-option label="正常" value="active" />
            <el-option label="待续费" value="past_due" />
            <el-option label="已停用" value="suspended" />
            <el-option label="已过期" value="expired" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button @click="resetFilters">重置筛选</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="section-card overview-card">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-card shadow="never" class="summary-card">
            <div class="summary-topline">机构总数</div>
            <div class="summary-value">{{ filteredOrganizations.length }}</div>
            <div class="summary-desc">当前筛选范围内机构数量</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="summary-card">
            <div class="summary-topline">高风险机构</div>
            <div class="summary-value">{{ highRiskCount }}</div>
            <div class="summary-desc">优先关注续费、停用与配额风险</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="summary-card">
            <div class="summary-topline">待分配负责人</div>
            <div class="summary-value">{{ unassignedFollowUpCount }}</div>
            <div class="summary-desc">建议尽快指定续费/运营负责人</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="summary-card">
            <div class="summary-topline">试用中机构</div>
            <div class="summary-value">{{ trialCount }}</div>
            <div class="summary-desc">重点推进试用转正式</div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="never" class="section-card data-table-card" v-loading="loadingOrganizations">
      <template #header>
        <div class="card-title">机构列表</div>
      </template>
      <el-table :data="filteredOrganizations" border>
        <el-table-column prop="name" label="机构名称" min-width="220" />
        <el-table-column prop="status" label="机构状态" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
              {{ scope.row.status === 'active' ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="subscriptionStatus" label="订阅状态" width="140">
          <template #default="scope">
            <el-tag :type="subscriptionTagType(scope.row.subscriptionStatus)">
              {{ formatSubscriptionStatus(scope.row.subscriptionStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="套餐版本" min-width="140">
          <template #default="scope">{{ formatPlanName(scope.row.planCode) }}</template>
        </el-table-column>
        <el-table-column label="到期时间" min-width="200">
          <template #default="scope">
            <span>{{ formatEndsAt(scope.row.endsAt) }}</span>
            <span v-if="scope.row.daysRemaining" class="days-remaining">（剩余 {{ scope.row.daysRemaining }} 天）</span>
          </template>
        </el-table-column>
        <el-table-column label="运营状态" min-width="220">
          <template #default="scope">
            <div class="meta-lines">
              <div>风险级别：{{ formatHealthLevel(scope.row.healthLevel) }}</div>
              <div>跟进状态：{{ scope.row.followUpStatus || '未设置' }}</div>
              <div>负责人：{{ scope.row.followUpOwner || '未分配' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="资源配额" min-width="220">
          <template #default="scope">
            <span>校区 {{ scope.row.usedCampuses || 0 }}/{{ scope.row.maxCampuses || 0 }}，账号 {{ scope.row.usedUsers || 0 }}/{{ scope.row.maxUsers || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="已开通功能" min-width="280">
          <template #default="scope">
            <el-space wrap>
              <el-tag v-for="flag in scope.row.featureFlags || []" :key="flag" size="small" effect="light">{{ formatFeatureLabel(flag) }}</el-tag>
            </el-space>
          </template>
        </el-table-column>
        <el-table-column label="系统信息" min-width="220">
          <template #default="scope">
            <div class="meta-lines">
              <div>机构编号：{{ scope.row.code || '-' }}</div>
              <div>机构ID：{{ scope.row.id || '-' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <el-space wrap>
              <el-button v-if="canUpdateTenant" link type="primary" @click="openOrgEdit(scope.row)">编辑</el-button>
              <el-button link @click="enterOrgView(scope.row)">进入机构视角</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="orgDialogVisible" :title="orgEditId ? '编辑机构' : '新增机构'" width="560px">
      <el-form label-position="top" :model="orgForm">
        <el-form-item label="机构编号">
          <el-input :model-value="orgEditId ? orgForm.orgCode : '系统自动生成'" disabled placeholder="保存后自动生成机构编号" />
        </el-form-item>
        <el-form-item label="机构名称">
          <el-input v-model="orgForm.orgName" />
        </el-form-item>
        <el-form-item label="机构状态">
          <el-select v-model="orgForm.status" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="订阅状态">
          <el-select v-model="orgForm.subscriptionStatus" style="width: 100%">
            <el-option label="试用中" value="trial" />
            <el-option label="正常" value="active" />
            <el-option label="待续费" value="past_due" />
            <el-option label="已停用" value="suspended" />
            <el-option label="已过期" value="expired" />
          </el-select>
        </el-form-item>
        <el-form-item label="套餐版本">
          <el-select v-model="orgForm.planCode" style="width: 100%">
            <el-option v-for="item in planOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="已绑定功能">
          <el-space wrap>
            <el-tag v-for="flag in selectedPlanFeatures" :key="flag" size="small" effect="light">{{ formatFeatureLabel(flag) }}</el-tag>
          </el-space>
        </el-form-item>
        <el-form-item label="到期时间">
          <el-input v-model="orgForm.endsAt" placeholder="2026-12-31 或 2026-12-31 23:59:59" />
        </el-form-item>
        <el-form-item label="最大校区数">
          <el-input-number v-model="orgForm.maxCampuses" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="最大账号数">
          <el-input-number v-model="orgForm.maxUsers" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="时区">
          <el-input v-model="orgForm.timezone" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="orgForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orgDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitOrganization">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createPlatformOrganization, fetchPlatformOrganizations, updatePlatformOrganization, type PlatformOrganizationItem, type PlatformOrganizationPayload } from '@/api/platform'
import { useAuthStore } from '@/stores/auth'
import { usePermission } from '@/composables/usePermission'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const authStore = useAuthStore()
const { can } = usePermission()
const router = useRouter()
const canCreateTenant = computed(() => can({ permission: 'platform:tenant:create', menuPermission: 'menu:platform:orgs:view' }))
const canUpdateTenant = computed(() => can({ permission: 'platform:tenant:update', menuPermission: 'menu:platform:orgs:view' }))
const loadingOrganizations = ref(false)
const submitting = ref(false)
const orgDialogVisible = ref(false)
const orgEditId = ref<number | null>(null)
const organizations = ref<PlatformOrganizationItem[]>([])
const filters = reactive({
  keyword: '',
  status: '',
  subscriptionStatus: '',
})

const planOptions = [
  { label: '试用版', value: 'trial' },
  { label: '标准版', value: 'standard' },
  { label: '专业版', value: 'pro' },
  { label: '旗舰版', value: 'enterprise' },
  { label: '平台版', value: 'platform' },
]

const orgForm = reactive<PlatformOrganizationPayload & { orgCode: string }>({
  orgCode: '',
  orgName: '',
  status: 'active',
  subscriptionStatus: 'trial',
  planCode: 'standard',
  endsAt: '',
  maxCampuses: 10,
  maxUsers: 500,
  timezone: 'Asia/Shanghai',
  remark: '',
})

const featureLabelMap: Record<string, string> = {
  auth: '认证登录',
  user_management: '用户管理',
  student: '学员管理',
  course: '课程管理',
  lesson_package: '课时包管理',
  payment: '收费管理',
  schedule: '排课管理',
  lesson_record: '上课记录',
  reschedule: '调补课管理',
  platform_management: '平台管理',
}

const planNameMap: Record<string, string> = {
  trial: '试用版',
  standard: '标准版',
  pro: '专业版',
  enterprise: '旗舰版',
  platform: '平台版',
}

const planFeatureMap: Record<string, string[]> = {
  trial: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule'],
  standard: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule'],
  pro: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule'],
  enterprise: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule', 'platform_management'],
  platform: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule', 'platform_management'],
}

const selectedPlanFeatures = computed(() => planFeatureMap[orgForm.planCode || 'standard'] || [])
const filteredOrganizations = computed(() => {
  const keyword = filters.keyword.trim().toLowerCase()
  return organizations.value.filter((item) => {
    const matchKeyword = !keyword || item.name?.toLowerCase().includes(keyword) || item.code?.toLowerCase().includes(keyword)
    const matchStatus = !filters.status || item.status === filters.status
    const matchSubscriptionStatus = !filters.subscriptionStatus || item.subscriptionStatus === filters.subscriptionStatus
    return matchKeyword && matchStatus && matchSubscriptionStatus
  })
})
const highRiskCount = computed(() => filteredOrganizations.value.filter((item) => item.healthLevel === 'high').length)
const unassignedFollowUpCount = computed(() => filteredOrganizations.value.filter((item) => (item.followUpOwner || '').trim() === '').length)
const trialCount = computed(() => filteredOrganizations.value.filter((item) => item.subscriptionStatus === 'trial').length)

onMounted(() => {
  loadOrganizations()
})

function formatFeatureLabel(flag?: string) {
  if (!flag) return '-'
  return featureLabelMap[flag] || flag
}

function formatPlanName(planCode?: string) {
  if (!planCode) return '未设置'
  return planNameMap[planCode] || planCode
}

function subscriptionTagType(status?: string) {
  if (status === 'active' || status === 'trial') return 'success'
  if (status === 'past_due') return 'warning'
  return 'danger'
}

function formatSubscriptionStatus(status?: string) {
  if (status === 'trial') return '试用中'
  if (status === 'active') return '正常'
  if (status === 'past_due') return '待续费'
  if (status === 'suspended') return '已停用'
  if (status === 'expired') return '已过期'
  return status || '-'
}

function formatEndsAt(value?: string) {
  return value || '未设置'
}

function formatHealthLevel(level?: string) {
  if (level === 'high') return '高风险'
  if (level === 'medium') return '需关注'
  return '健康'
}

async function loadOrganizations() {
  loadingOrganizations.value = true
  try {
    const resp = await fetchPlatformOrganizations()
    organizations.value = resp.data?.data || []
  } catch (error) {
    message.error(extractErrorMessage(error, '加载机构数据失败，请稍后重试'))
  } finally {
    loadingOrganizations.value = false
  }
}

function resetFilters() {
  filters.keyword = ''
  filters.status = ''
  filters.subscriptionStatus = ''
}

function resetOrgForm() {
  orgForm.orgCode = ''
  orgForm.orgName = ''
  orgForm.status = 'active'
  orgForm.subscriptionStatus = 'trial'
  orgForm.planCode = 'standard'
  orgForm.endsAt = ''
  orgForm.maxCampuses = 10
  orgForm.maxUsers = 500
  orgForm.timezone = 'Asia/Shanghai'
  orgForm.remark = ''
}

function openOrgCreate() {
  if (!canCreateTenant.value) {
    message.warning('没有新增机构权限')
    return
  }
  orgEditId.value = null
  resetOrgForm()
  orgDialogVisible.value = true
}

function openOrgEdit(item: PlatformOrganizationItem) {
  if (!canUpdateTenant.value) {
    message.warning('没有编辑机构权限')
    return
  }
  orgEditId.value = item.id
  orgForm.orgCode = item.code
  orgForm.orgName = item.name
  orgForm.status = item.status
  orgForm.subscriptionStatus = item.subscriptionStatus || 'active'
  orgForm.planCode = item.planCode || item.editionCode || 'standard'
  orgForm.endsAt = item.endsAt || ''
  orgForm.maxCampuses = item.maxCampuses || 10
  orgForm.maxUsers = item.maxUsers || 500
  orgForm.timezone = 'Asia/Shanghai'
  orgForm.remark = ''
  orgDialogVisible.value = true
}

function enterOrgView(item: PlatformOrganizationItem) {
  authStore.setPlatformViewAs({
    organizationId: item.id,
    organizationName: item.name,
  })
  message.success(`已进入 ${item.name} 的机构视角`)
  void router.push('/platform/users')
}

async function submitOrganization() {
  if (orgEditId.value && !canUpdateTenant.value) {
    message.warning('没有编辑机构权限')
    return
  }
  if (!orgEditId.value && !canCreateTenant.value) {
    message.warning('没有新增机构权限')
    return
  }
  if (!orgForm.orgName.trim()) {
    message.warning('请先填写机构名称')
    return
  }
  submitting.value = true
  try {
    const payload = {
      orgName: orgForm.orgName.trim(),
      status: orgForm.status,
      subscriptionStatus: orgForm.subscriptionStatus,
      planCode: orgForm.planCode?.trim(),
      endsAt: orgForm.endsAt?.trim(),
      maxCampuses: orgForm.maxCampuses,
      maxUsers: orgForm.maxUsers,
      timezone: orgForm.timezone?.trim(),
      remark: orgForm.remark?.trim(),
    }
    if (orgEditId.value) {
      await updatePlatformOrganization(orgEditId.value, payload)
      message.success('机构更新成功')
    } else {
      await createPlatformOrganization(payload)
      message.success('机构创建成功')
    }
    orgDialogVisible.value = false
    await loadOrganizations()
  } catch (error) {
    message.error(extractErrorMessage(error, '保存机构失败，请稍后重试'))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page-card { display: grid; gap: 16px; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; }
.hero-kicker { font-size: 12px; font-weight: 700; color: #2563eb; letter-spacing: 0.08em; text-transform: uppercase; }
.card-title { font-weight: 600; }
.actions { display: flex; gap: 12px; flex-wrap: wrap; }
.days-remaining { color: #2563eb; font-size: 12px; }
.meta-lines { display: grid; gap: 4px; color: #606266; font-size: 12px; }
</style>
