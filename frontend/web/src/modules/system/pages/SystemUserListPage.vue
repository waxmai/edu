<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">安全与权限</div>
        <h2>系统用户</h2>
        <p>统一维护账号、角色、状态与登录安全治理。</p>
      </div>
      <div class="actions">
        <el-button @click="loadUsers">刷新</el-button>
        <el-button v-if="canCreateUser" type="primary" @click="openCreate">新增用户</el-button>
      </div>
    </div>

    <el-alert
      title="仅管理员可维护系统用户，支持启用、禁用、锁定、解锁与重置密码。"
      type="success"
      :closable="false"
    />

    <el-card shadow="never" class="toolbar-card search-form-card">
      <el-form :inline="true" :model="query" @submit.prevent>
        <el-form-item label="用户名">
          <el-input v-model="query.username" placeholder="按用户名搜索" clearable />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="query.realName" placeholder="按姓名搜索" clearable />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="query.roleCode" placeholder="全部角色" clearable style="width: 160px">
            <el-option v-for="item in roleOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属机构" v-if="isPlatformAdmin">
          <el-select v-model="query.organizationId" placeholder="全部机构" clearable filterable style="width: 220px">
            <el-option v-for="item in organizations" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属校区" v-if="isPlatformAdmin">
          <el-select v-model="query.campusId" placeholder="全部校区" clearable filterable style="width: 220px">
            <el-option v-for="item in availableCampuses" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="全部状态" clearable style="width: 140px">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
            <el-option label="锁定" value="locked" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadUsers">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="section-card data-table-card" v-loading="loading">
      <el-table :data="users" border>
        <el-table-column prop="username" label="用户名" min-width="150" />
        <el-table-column prop="realName" label="姓名" min-width="150" />
        <el-table-column prop="roleCode" label="角色" min-width="140">
          <template #default="scope">
            <el-tag :type="roleTagType(scope.row.roleCode)" effect="light" round class="status-tag">
              {{ formatRole(scope.row.roleCode) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="归属范围" min-width="220">
          <template #default="scope">{{ formatScope(scope.row) }}</template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="160" />
        <el-table-column prop="status" label="状态" min-width="140">
          <template #default="scope">
            <el-tag :type="statusTagType(scope.row.status)">{{ formatStatus(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="failedLoginCount" label="失败次数" width="100" />
        <el-table-column prop="lockedUntil" label="锁定至" min-width="180">
          <template #default="scope">{{ scope.row.lockedUntil || '-' }}</template>
        </el-table-column>
        <el-table-column label="系统信息" min-width="220">
          <template #default="scope">
            <div class="meta-lines">
              <div>用户ID：{{ scope.row.id || '-' }}</div>
              <div>机构：{{ formatOrganizationName(scope.row.organizationId) }}</div>
              <div>校区：{{ formatCampusName(scope.row.campusId) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="180" show-overflow-tooltip />
        <el-table-column prop="updatedAt" label="更新时间" min-width="180">
          <template #default="scope">{{ scope.row.updatedAt || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="360" fixed="right">
          <template #default="scope">
            <el-space wrap>
              <el-button v-if="canEditUser" link type="primary" @click="openEdit(scope.row)">编辑</el-button>
              <el-button v-if="canChangeUserStatus" link @click="changeStatus(scope.row, 'enabled')">启用</el-button>
              <el-button v-if="canChangeUserStatus" link @click="changeStatus(scope.row, 'disabled')">禁用</el-button>
              <el-button v-if="canChangeUserStatus" link type="danger" @click="changeStatus(scope.row, 'locked')">锁定</el-button>
              <el-button v-if="canUnlockUser" link type="warning" @click="unlock(scope.row)">解锁</el-button>
              <el-button v-if="canResetUserPassword" link @click="resetPassword(scope.row)">重置密码</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && users.length === 0" description="还没有系统用户，可先创建第一个账号" />

      <div class="pagination-wrap">
        <el-pagination
          background
          layout="total, prev, pager, next, sizes"
          :current-page="query.pageNum"
          :page-size="query.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="520px">
      <el-form label-position="top" :model="form">
        <el-form-item label="用户名">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item :label="isEdit ? '密码（留空则不修改）' : '密码（新增时必填，至少12位且需包含大小写字母、数字、特殊字符）'">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model="form.realName" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.roleCode" style="width: 100%">
            <el-option v-for="item in roleOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="requiresOrganization" label="所属机构">
          <el-select v-model="form.organizationId" style="width: 100%" clearable filterable>
            <el-option v-for="item in organizations" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="requiresCampus" label="所属校区">
          <el-select v-model="form.campusId" style="width: 100%" :disabled="!form.organizationId" clearable filterable>
            <el-option v-for="item in formCampusOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
            <el-option label="锁定" value="locked" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  createSystemUser,
  fetchSystemUsers,
  resetSystemUserPassword,
  unlockSystemUser,
  updateSystemUser,
  updateSystemUserStatus,
  type SystemUserFormPayload,
  type SystemUserItem,
} from '@/api/system'
import { fetchPlatformCampuses, fetchPlatformOrganizations, type PlatformCampusItem, type PlatformOrganizationItem } from '@/api/platform'
import { usePlatformViewScope } from '@/composables/usePlatformViewScope'
import { useAuthStore } from '@/stores/auth'
import { usePermission } from '@/composables/usePermission'
import { confirmDialog } from '@/utils/confirm'
import { extractErrorMessage, isCancelError } from '@/utils/error'
import { message } from '@/utils/message'
import { cleanQueryParams } from '@/utils/query'

const authStore = useAuthStore()
const { can } = usePermission()
const { withPlatformViewScope } = usePlatformViewScope()
const isPlatformAdmin = computed(() => authStore.currentUser?.roleCode === 'platform_admin')
const canCreateUser = computed(() => can({ permission: 'system:user:create', menuPermission: 'menu:system:user:view' }))
const canEditUser = computed(() => can({ permission: 'system:user:update', menuPermission: 'menu:system:user:view' }))
const canChangeUserStatus = computed(() => can({ permission: 'system:user:status', menuPermission: 'menu:system:user:view' }))
const canUnlockUser = computed(() => can({ permission: 'system:user:unlock', menuPermission: 'menu:system:user:view' }))
const canResetUserPassword = computed(() => can({ permission: 'system:user:reset-password', menuPermission: 'menu:system:user:view' }))
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const users = ref<SystemUserItem[]>([])
const total = ref(0)
const organizations = ref<PlatformOrganizationItem[]>([])
const campuses = ref<PlatformCampusItem[]>([])

const roleOptions = [
  { label: '平台超管', value: 'platform_admin' },
  { label: '机构超管', value: 'org_admin' },
  { label: '校区管理员', value: 'campus_admin' },
  { label: '教师', value: 'teacher' },
]

const query = reactive({
  username: '',
  realName: '',
  roleCode: '',
  organizationId: undefined as number | undefined,
  campusId: undefined as number | undefined,
  status: '',
  pageNum: 1,
  pageSize: 10,
})

const availableCampuses = computed(() => {
  if (!isPlatformAdmin.value || !query.organizationId) {
    return campuses.value
  }
  return campuses.value.filter((item) => item.organizationId === query.organizationId)
})
const requiresOrganization = computed(() => form.roleCode !== 'platform_admin')
const requiresCampus = computed(() => form.roleCode === 'campus_admin' || form.roleCode === 'teacher')
const formCampusOptions = computed(() => {
  if (!form.organizationId) {
    return []
  }
  return campuses.value.filter((item) => item.organizationId === form.organizationId)
})

const form = reactive<SystemUserFormPayload>({
  username: '',
  password: '',
  roleCode: 'teacher',
  organizationId: 1,
  campusId: 1,
  realName: '',
  phone: '',
  status: 'enabled',
  remark: '',
})

onMounted(() => {
  loadScopeOptions()
  loadUsers()
})

watch(() => query.organizationId, () => {
  query.campusId = undefined
})

watch(() => form.organizationId, () => {
  if (requiresCampus.value) {
    form.campusId = undefined
  }
})

watch(() => form.roleCode, (roleCode) => {
  if (roleCode === 'platform_admin') {
    form.organizationId = undefined
    form.campusId = undefined
    return
  }
  if (roleCode === 'org_admin') {
    form.campusId = undefined
  }
})

function resetForm() {
  form.username = ''
  form.password = ''
  form.roleCode = 'teacher'
  form.organizationId = isPlatformAdmin.value && authStore.platformViewAs?.organizationId ? authStore.platformViewAs.organizationId : undefined
  form.campusId = isPlatformAdmin.value && authStore.platformViewAs?.campusId ? authStore.platformViewAs.campusId : undefined
  form.realName = ''
  form.phone = ''
  form.status = 'enabled'
  form.remark = ''
}

function openCreate() {
  if (!canCreateUser.value) {
    message.warning('没有新增用户权限')
    return
  }
  isEdit.value = false
  editingId.value = null
  resetForm()
  if (isPlatformAdmin.value && authStore.platformViewAs?.organizationId) {
    form.organizationId = authStore.platformViewAs.organizationId
    form.campusId = authStore.platformViewAs.campusId || undefined
  }
  dialogVisible.value = true
}

function openEdit(user: SystemUserItem) {
  if (!canEditUser.value) {
    message.warning('没有编辑用户权限')
    return
  }
  isEdit.value = true
  editingId.value = user.id
  form.username = user.username
  form.password = ''
  form.roleCode = user.roleCode
  form.organizationId = user.organizationId || undefined
  form.campusId = user.campusId || undefined
  form.realName = user.realName
  form.phone = user.phone || ''
  form.status = user.status
  form.remark = user.remark || ''
  dialogVisible.value = true
}

function resetQuery() {
  query.username = ''
  query.realName = ''
  query.roleCode = ''
  query.organizationId = undefined
  query.campusId = undefined
  query.status = ''
  query.pageNum = 1
  query.pageSize = 10
  loadUsers()
}

function statusTagType(status: string) {
  if (status === 'enabled') return 'success'
  if (status === 'disabled') return 'warning'
  return 'danger'
}

function roleTagType(roleCode: string) {
  if (roleCode === 'platform_admin') return 'danger'
  if (roleCode === 'org_admin') return 'warning'
  if (roleCode === 'campus_admin') return 'primary'
  return 'success'
}

function formatStatus(status: string) {
  if (status === 'enabled') return '启用'
  if (status === 'disabled') return '禁用'
  if (status === 'locked') return '锁定'
  return status || '-'
}

function formatRole(roleCode: string) {
  if (roleCode === 'platform_admin') return '平台超管'
  if (roleCode === 'org_admin') return '机构超管'
  if (roleCode === 'campus_admin') return '校区管理员'
  if (roleCode === 'teacher') return '教师'
  return roleCode || '-'
}

function formatScope(user: SystemUserItem) {
  if (user.roleCode === 'platform_admin') return '平台级'
  if (user.roleCode === 'org_admin') return user.organizationId ? `机构级（${formatOrganizationName(user.organizationId)}）` : '机构级'
  if (user.roleCode === 'campus_admin') return user.campusId ? `校区级（${formatCampusName(user.campusId)}）` : '校区级'
  if (user.roleCode === 'teacher') return user.campusId ? `校区教师（${formatCampusName(user.campusId)}）` : '教师'
  return '-'
}

function formatOrganizationName(organizationId?: number) {
  return organizations.value.find((item) => item.id === organizationId)?.name || '-'
}

function formatCampusName(campusId?: number) {
  return campuses.value.find((item) => item.id === campusId)?.name || '-'
}

async function loadScopeOptions() {
  if (!isPlatformAdmin.value) {
    return
  }
  try {
    const [orgResp, campusResp] = await Promise.all([fetchPlatformOrganizations(), fetchPlatformCampuses()])
    organizations.value = orgResp.data?.data || []
    campuses.value = campusResp.data?.data || []
  } catch (error) {
    message.error(extractErrorMessage(error, '加载机构/校区选项失败，请稍后重试'))
  }
}

async function loadUsers() {
  loading.value = true
  try {
    const response = await fetchSystemUsers(cleanQueryParams(withPlatformViewScope({
      ...query,
    })))
    users.value = response.data?.data?.list || []
    total.value = response.data?.data?.total || 0
  } catch (error) {
    console.error(error)
    message.error(extractErrorMessage(error, '获取用户列表失败，请稍后重试'))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  query.pageNum = page
  loadUsers()
}

function handleSizeChange(size: number) {
  query.pageSize = size
  query.pageNum = 1
  loadUsers()
}

async function submitForm() {
  if (isEdit.value && !canEditUser.value) {
    message.warning('没有编辑用户权限')
    return
  }
  if (!isEdit.value && !canCreateUser.value) {
    message.warning('没有新增用户权限')
    return
  }
  if (!form.username.trim() || !form.realName.trim() || !form.roleCode.trim()) {
    message.warning('请先填写用户名、姓名和角色')
    return
  }
  if (!isEdit.value && !form.password?.trim()) {
    form.password = ''
  }

  submitting.value = true
  try {
    const payload: SystemUserFormPayload = {
      username: form.username.trim(),
      password: form.password?.trim(),
      roleCode: form.roleCode,
      organizationId: requiresOrganization.value ? form.organizationId : undefined,
      campusId: requiresCampus.value ? form.campusId : undefined,
      realName: form.realName.trim(),
      phone: form.phone?.trim(),
      status: form.status,
      remark: form.remark?.trim(),
    }

    if (isEdit.value && editingId.value) {
      await updateSystemUser(editingId.value, payload)
      message.success('用户更新成功')
    } else {
      await createSystemUser(payload)
      message.success('用户创建成功')
    }

    dialogVisible.value = false
    await loadUsers()
  } catch (error) {
    console.error(error)
    message.error(extractErrorMessage(error, '保存失败，请检查用户名、角色和手机号后重试'))
  } finally {
    submitting.value = false
  }
}

async function changeStatus(user: SystemUserItem, status: string) {
  if (!canChangeUserStatus.value) {
    message.warning('没有更新用户状态权限')
    return
  }
  const actionText = status === 'enabled' ? '启用' : status === 'disabled' ? '禁用' : '锁定'
  try {
    await confirmDialog(`确认${actionText}用户 ${user.username} 吗？`, `${actionText}确认`, { type: 'warning' })
    await updateSystemUserStatus(user.id, status)
    message.success('状态更新成功')
    await loadUsers()
  } catch (error) {
    if (!isCancelError(error)) {
      console.error(error)
      message.error(extractErrorMessage(error, '状态更新失败，请稍后重试'))
    }
  }
}

async function unlock(user: SystemUserItem) {
  if (!canUnlockUser.value) {
    message.warning('没有解锁用户权限')
    return
  }
  try {
    await confirmDialog(`确认解锁用户 ${user.username} 吗？`, '解锁确认', { type: 'warning' })
    await unlockSystemUser(user.id)
    message.success('用户已解锁')
    await loadUsers()
  } catch (error) {
    if (!isCancelError(error)) {
      console.error(error)
      message.error(extractErrorMessage(error, '用户解锁失败，请稍后重试'))
    }
  }
}

async function resetPassword(user: SystemUserItem) {
  if (!canResetUserPassword.value) {
    message.warning('没有重置密码权限')
    return
  }
  try {
    await confirmDialog(`确认重置用户 ${user.username} 的密码吗？默认会要求首次改密。`, '重置密码确认', { type: 'warning' })
    await resetSystemUserPassword(user.id)
    message.success('密码已重置为默认口令，请通知用户尽快改密')
    await loadUsers()
  } catch (error) {
    if (!isCancelError(error)) {
      console.error(error)
      message.error(extractErrorMessage(error, '密码重置失败，请稍后重试'))
    }
  }
}
</script>

<style scoped>
.page-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.actions {
  display: flex;
  gap: 12px;
}

.meta-lines {
  display: grid;
  gap: 4px;
  color: #606266;
  font-size: 12px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
