<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">校区账号治理</div>
        <h2>校区用户管理</h2>
        <p>维护本校区教师账号与现场使用账号状态。</p>
      </div>
      <div class="actions">
        <el-button @click="loadUsers">刷新</el-button>
        <el-button type="primary" @click="openCreate">新增用户</el-button>
      </div>
    </div>

    <el-alert
      title="当前页面聚焦本校区教师与现场执行账号维护。"
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
        <el-table-column label="归属范围" min-width="180">
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
        <el-table-column label="系统信息" min-width="180">
          <template #default="scope">
            <div class="meta-lines">
              <div>用户ID：{{ scope.row.id || '-' }}</div>
              <div>机构ID：{{ scope.row.organizationId || '-' }}</div>
              <div>校区ID：{{ scope.row.campusId || '-' }}</div>
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
              <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
              <el-button link @click="changeStatus(scope.row, 'enabled')">启用</el-button>
              <el-button link @click="changeStatus(scope.row, 'disabled')">禁用</el-button>
              <el-button link type="danger" @click="changeStatus(scope.row, 'locked')">锁定</el-button>
              <el-button link type="warning" @click="unlock(scope.row)">解锁</el-button>
              <el-button link @click="resetPassword(scope.row)">重置密码</el-button>
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
        <el-form-item label="机构">
          <el-input :model-value="authStore.currentUser?.organizationName || `机构 #${currentOrganizationId || '-'}`" disabled />
        </el-form-item>
        <el-form-item label="校区">
          <el-input :model-value="authStore.currentUser?.campusName || formatCampusName(currentCampusId)" disabled />
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
import { computed, onMounted, reactive, ref } from 'vue'
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
import { fetchPlatformCampuses, type PlatformCampusItem } from '@/api/platform'
import { useAuthStore } from '@/stores/auth'
import { confirmDialog } from '@/utils/confirm'
import { extractErrorMessage, isCancelError } from '@/utils/error'
import { message } from '@/utils/message'
import { cleanQueryParams } from '@/utils/query'

const authStore = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const users = ref<SystemUserItem[]>([])
const total = ref(0)
const campusOptions = ref<PlatformCampusItem[]>([])

const roleOptions = [
  { label: '教师', value: 'teacher' },
]

const currentOrganizationId = computed(() => authStore.currentUser?.organizationId)
const currentCampusId = computed(() => authStore.currentUser?.campusId)

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
  loadCampusOptions()
  loadUsers()
})

function resetForm() {
  form.username = ''
  form.password = ''
  form.roleCode = 'teacher'
  form.organizationId = currentOrganizationId.value
  form.campusId = currentCampusId.value
  form.realName = ''
  form.phone = ''
  form.status = 'enabled'
  form.remark = ''
}

function openCreate() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(user: SystemUserItem) {
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
  return roleCode === 'teacher' ? 'success' : 'primary'
}

function formatStatus(status: string) {
  if (status === 'enabled') return '启用'
  if (status === 'disabled') return '禁用'
  if (status === 'locked') return '锁定'
  return status || '-'
}

function formatRole(roleCode: string) {
  if (roleCode === 'teacher') return '教师'
  return roleCode || '-'
}

function formatScope(user: SystemUserItem) {
  if (user.roleCode === 'teacher') return user.campusId ? `校区教师（${formatCampusName(user.campusId)}）` : '教师'
  return '-'
}

function formatCampusName(campusId?: number) {
  return campusOptions.value.find((item) => item.id === campusId)?.name || (campusId ? `校区 #${campusId}` : '-')
}

async function loadCampusOptions() {
  try {
    const resp = await fetchPlatformCampuses()
    campusOptions.value = resp.data?.data || []
  } catch (error) {
    message.error(extractErrorMessage(error, '加载校区选项失败，请稍后重试'))
  }
}

async function loadUsers() {
  loading.value = true
  try {
    const response = await fetchSystemUsers(cleanQueryParams({
      ...query,
      roleCode: query.roleCode || 'teacher',
      organizationId: currentOrganizationId.value,
      campusId: currentCampusId.value,
    }))
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
      organizationId: currentOrganizationId.value,
      campusId: currentCampusId.value,
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
