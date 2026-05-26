<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">{{ isPlatformAdmin ? 'SaaS 平台治理' : '机构运营配置' }}</div>
        <h2>{{ isPlatformAdmin ? '校区管理' : '校区管理' }}</h2>
        <p>{{ isPlatformAdmin ? '统一维护校区组织结构、归属关系与基础启停状态。' : '维护本机构校区结构与可用状态，支撑机构日常运营。' }}</p>
      </div>
      <div class="actions">
        <el-button @click="loadData">刷新</el-button>
        <el-button v-if="canCreateCampus" type="success" @click="openCampusCreate">新增校区</el-button>
      </div>
    </div>

    <el-alert
      :title="isPlatformAdmin ? '平台管理员可在此维护校区基础信息，并可进入校区视角排障。' : '机构管理员可在本机构范围内维护校区基础信息，支持新增和编辑校区。'"
      type="warning"
      :closable="false"
    />

    <el-card v-if="isPlatformAdmin" shadow="never" class="section-card filter-card">
      <el-form inline>
        <el-form-item label="所属机构">
          <el-select v-model="filters.organizationId" placeholder="全部机构" clearable filterable style="width: 220px">
            <el-option v-for="item in organizations" :key="item.id" :label="item.name" :value="item.id" />
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
            <div class="summary-topline">校区总数</div>
            <div class="summary-value">{{ filteredCampuses.length }}</div>
            <div class="summary-desc">当前机构范围内可管理校区</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="summary-card">
            <div class="summary-topline">启用校区</div>
            <div class="summary-value">{{ activeCampusCount }}</div>
            <div class="summary-desc">状态为启用的校区数量</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="summary-card">
            <div class="summary-topline">启用账号数</div>
            <div class="summary-value">{{ totalActiveUsers }}</div>
            <div class="summary-desc">各校区启用账号合计</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="summary-card">
            <div class="summary-topline">异常总数</div>
            <div class="summary-value">{{ totalRiskCount }}</div>
            <div class="summary-desc">用于识别优先治理校区</div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="never" class="section-card data-table-card" v-loading="loadingCampuses">
      <template #header>
        <div class="card-title">校区列表</div>
      </template>
      <el-table :data="sortedCampuses" border>
        <el-table-column prop="name" label="校区名称" min-width="200" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
              {{ scope.row.status === 'active' ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="治理概况" min-width="220">
          <template #default="scope">
            <div class="meta-lines">
              <div>启用账号：{{ scope.row.usedUsers || 0 }}</div>
              <div>异常总数：{{ scope.row.daysRemaining || 0 }}</div>
              <div>套餐：{{ formatPlanName(scope.row.planCode) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="优先级" width="120">
          <template #default="scope">
            <el-tag :type="riskTagType(scope.row.healthLevel)">{{ riskPriorityLabel(scope.row.healthLevel) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="已开通功能" min-width="260">
          <template #default="scope">
            <el-space wrap>
              <el-tag v-for="flag in scope.row.featureFlags || []" :key="flag" size="small" effect="light">{{ formatFeatureLabel(flag) }}</el-tag>
            </el-space>
          </template>
        </el-table-column>
        <el-table-column label="系统信息" min-width="260">
          <template #default="scope">
            <div class="meta-lines">
              <div>所属机构：{{ formatOrganizationName(scope.row.organizationId) }}</div>
              <div>所属机构ID：{{ scope.row.organizationId || '-' }}</div>
              <div>校区编号：{{ scope.row.code || '-' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="scope">
            <el-space wrap>
              <el-button v-if="canUpdateCampus" link type="primary" @click="openCampusEdit(scope.row)">编辑</el-button>
              <el-button v-if="isPlatformAdmin" link type="success" @click="enterCampusView(scope.row)">进入校区视角</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="campusDialogVisible" :title="campusEditId ? '编辑校区' : '新增校区'" width="560px">
      <el-form label-position="top" :model="campusForm">
        <el-form-item label="所属机构">
          <el-select v-model="campusForm.organizationId" :disabled="!isPlatformAdmin || Boolean(campusEditId)" style="width: 100%">
            <el-option v-for="item in organizations" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="校区编号">
          <el-input :model-value="campusEditId ? campusForm.campusCode : '系统自动生成'" disabled placeholder="保存后自动生成校区编号" />
        </el-form-item>
        <el-form-item label="校区名称">
          <el-input v-model="campusForm.campusName" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="campusForm.status" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="已开通功能">
          <el-space wrap>
            <el-tag v-for="flag in selectedCampusFeatures" :key="flag" size="small" effect="light">{{ formatFeatureLabel(flag) }}</el-tag>
          </el-space>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="campusForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="campusDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitCampus">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createPlatformCampus, fetchPlatformCampuses, fetchPlatformOrganizations, updatePlatformCampus, type PlatformCampusItem, type PlatformCampusPayload, type PlatformOrganizationItem } from '@/api/platform'
import { useAuthStore } from '@/stores/auth'
import { usePermission } from '@/composables/usePermission'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const authStore = useAuthStore()
const { can } = usePermission()
const router = useRouter()
const isPlatformAdmin = computed(() => authStore.currentUser?.roleCode === 'platform_admin')
const canCreateCampus = computed(() => can({ permission: 'platform:campus:create', menuPermission: 'menu:campus:manage:view' }))
const canUpdateCampus = computed(() => can({ permission: 'platform:campus:update', menuPermission: 'menu:campus:manage:view' }))
const loadingCampuses = ref(false)
const submitting = ref(false)
const campusDialogVisible = ref(false)
const campusEditId = ref<number | null>(null)
const organizations = ref<PlatformOrganizationItem[]>([])
const campuses = ref<PlatformCampusItem[]>([])
const filters = reactive({
  organizationId: undefined as number | undefined,
})

const campusForm = reactive<PlatformCampusPayload & { campusCode: string }>({
  organizationId: 1,
  campusCode: '',
  campusName: '',
  status: 'active',
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

const planFeatureMap: Record<string, string[]> = {
  trial: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule'],
  standard: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule'],
  pro: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule'],
  enterprise: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule', 'platform_management'],
  platform: ['auth', 'user_management', 'student', 'course', 'lesson_package', 'payment', 'schedule', 'lesson_record', 'reschedule', 'platform_management'],
}

const planNameMap: Record<string, string> = {
  trial: '试用版',
  standard: '标准版',
  pro: '专业版',
  enterprise: '旗舰版',
  platform: '平台版',
}

const selectedCampusFeatures = computed(() => {
  const orgId = Number(campusForm.organizationId || 0)
  const currentOrg = organizations.value.find((item) => item.id === orgId)
  const planCode = currentOrg?.planCode || currentOrg?.editionCode || 'standard'
  return planFeatureMap[planCode] || []
})
const filteredCampuses = computed(() => {
  if (!isPlatformAdmin.value || !filters.organizationId) {
    return campuses.value
  }
  return campuses.value.filter((item) => item.organizationId === filters.organizationId)
})
const sortedCampuses = computed(() => [...filteredCampuses.value].sort((a, b) => (b.daysRemaining || 0) - (a.daysRemaining || 0)))
const activeCampusCount = computed(() => filteredCampuses.value.filter((item) => item.status === 'active').length)
const totalActiveUsers = computed(() => filteredCampuses.value.reduce((sum, item) => sum + Number(item.usedUsers || 0), 0))
const totalRiskCount = computed(() => filteredCampuses.value.reduce((sum, item) => sum + Number(item.daysRemaining || 0), 0))

onMounted(() => {
  loadData()
})

function formatFeatureLabel(flag?: string) {
  if (!flag) return '-'
  return featureLabelMap[flag] || flag
}

function formatPlanName(planCode?: string) {
  if (!planCode) return '未设置'
  return planNameMap[planCode] || planCode
}

function riskPriorityLabel(level?: string) {
  if (level === 'high') return '高优先级'
  if (level === 'medium') return '需关注'
  return '稳定'
}

function riskTagType(level?: string) {
  if (level === 'high') return 'danger'
  if (level === 'medium') return 'warning'
  return 'success'
}

function formatOrganizationName(organizationId?: number) {
  return organizations.value.find((item) => item.id === organizationId)?.name || '-'
}

function resetFilters() {
  filters.organizationId = undefined
}

async function loadData() {
  loadingCampuses.value = true
  try {
    const campusResp = await fetchPlatformCampuses()
    campuses.value = campusResp.data?.data || []
    if (isPlatformAdmin.value) {
      const orgResp = await fetchPlatformOrganizations()
      organizations.value = orgResp.data?.data || []
    } else {
      const orgId = authStore.currentUser?.organizationId
      organizations.value = orgId
        ? [{
            id: orgId,
            code: '',
            name: authStore.currentUser?.organizationName || `机构 #${orgId}`,
            status: 'active',
            subscriptionStatus: authStore.currentUser?.subscriptionStatus,
            editionCode: 'standard',
            planCode: 'standard',
          }]
        : []
    }
  } catch (error) {
    message.error(extractErrorMessage(error, '加载校区数据失败，请稍后重试'))
  } finally {
    loadingCampuses.value = false
  }
}

function resetCampusForm() {
  campusForm.organizationId = isPlatformAdmin.value ? (organizations.value[0]?.id || 1) : (authStore.currentUser?.organizationId || 1)
  campusForm.campusCode = ''
  campusForm.campusName = ''
  campusForm.status = 'active'
  campusForm.remark = ''
}

function openCampusCreate() {
  if (!canCreateCampus.value) {
    message.warning('没有新增校区权限')
    return
  }
  campusEditId.value = null
  resetCampusForm()
  campusDialogVisible.value = true
}

function openCampusEdit(item: PlatformCampusItem) {
  if (!canUpdateCampus.value) {
    message.warning('没有编辑校区权限')
    return
  }
  campusEditId.value = item.id
  campusForm.organizationId = item.organizationId
  campusForm.campusCode = item.code
  campusForm.campusName = item.name
  campusForm.status = item.status
  campusForm.remark = ''
  campusDialogVisible.value = true
}

function enterCampusView(item: PlatformCampusItem) {
  if (!isPlatformAdmin.value) {
    message.info('机构管理员已处于本机构视角，无需切换校区视角')
    return
  }
  const orgName = formatOrganizationName(item.organizationId)
  authStore.setPlatformViewAs({
    organizationId: item.organizationId,
    organizationName: orgName === '-' ? undefined : orgName,
    campusId: item.id,
    campusName: item.name,
  })
  message.success(`已进入 ${item.name} 的校区视角`)
  void router.push('/platform/users')
}

async function submitCampus() {
  if (campusEditId.value && !canUpdateCampus.value) {
    message.warning('没有编辑校区权限')
    return
  }
  if (!campusEditId.value && !canCreateCampus.value) {
    message.warning('没有新增校区权限')
    return
  }
  if (!campusForm.campusName.trim() || !campusForm.organizationId) {
    message.warning('请先填写所属机构和校区名称')
    return
  }
  submitting.value = true
  try {
    const payload = {
      organizationId: campusForm.organizationId,
      campusName: campusForm.campusName.trim(),
      status: campusForm.status,
      remark: campusForm.remark?.trim(),
    }
    if (campusEditId.value) {
      await updatePlatformCampus(campusEditId.value, payload)
      message.success('校区更新成功')
    } else {
      await createPlatformCampus(payload)
      message.success('校区创建成功')
    }
    campusDialogVisible.value = false
    await loadData()
  } catch (error) {
    message.error(extractErrorMessage(error, '保存校区失败，请稍后重试'))
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
.meta-lines { display: grid; gap: 4px; color: #606266; font-size: 12px; }
</style>
