<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">课程资产</div>
        <h2>课程定义</h2>
        <p>维护课程基础信息，统一时长、收费与业务状态口径。</p>
      </div>
      <el-button v-if="canCreateCourse" type="primary" @click="openCreateDialog">新增课程</el-button>
    </div>

      <el-alert
        title="当前已接入正式课程定义接口，可维护课程名称、科目、类型、时长、收费和状态。"
        type="success"
        :closable="false"
      />

    <el-card shadow="never" class="section-card data-table-card">
      <el-table :data="rows" v-loading="loading" border>
        <el-table-column prop="courseName" label="课程名称" min-width="160" />
        <el-table-column prop="subject" label="科目" min-width="120" />
        <el-table-column prop="courseType" label="课程类型" min-width="120" />
        <el-table-column prop="durationMinutes" label="标准时长(分钟)" min-width="140" />
        <el-table-column prop="feeStandard" label="标准收费" min-width="120" />
        <el-table-column label="状态" min-width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'enabled' ? 'success' : 'info'" effect="light" round>
              {{ formatStatus(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" fixed="right" min-width="160">
          <template #default="scope">
            <el-space>
              <el-button v-if="canEditCourse" link type="primary" @click="openEditDialog(scope.row)">编辑</el-button>
              <el-button v-if="canDeleteCourse" link type="danger" @click="handleDelete(scope.row)">删除</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
      <div class="table-footer">
        <el-pagination
          background
          layout="total, prev, pager, next, sizes"
          :total="total"
          :page-size="query.pageSize || 10"
          :current-page="query.pageNum || 1"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑课程' : '新增课程'" width="640px">
      <el-form :model="form" label-width="110px">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="课程名称"><el-input v-model="form.courseName" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="科目"><el-input v-model="form.subject" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="课程类型"><el-input v-model="form.courseType" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="标准时长"><el-input-number v-model="form.durationMinutes" :min="1" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="标准收费"><el-input-number v-model="form.feeStandard" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="状态"><el-select v-model="form.status" style="width: 100%"><el-option label="启用" value="enabled" /><el-option label="禁用" value="disabled" /></el-select></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="3" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { createCourse, deleteCourse, fetchCourseList, updateCourse, type CourseFormModel, type CourseItem, type CourseSearchParams } from '@/api/course'
import { usePlatformViewScope } from '@/composables/usePlatformViewScope'
import { usePermission } from '@/composables/usePermission'
import { useAuthStore } from '@/stores/auth'
import { normalizeListPayload } from '@/utils/api'
import { confirmDialog } from '@/utils/confirm'
import { extractErrorMessage, isCancelError } from '@/utils/error'
import { message } from '@/utils/message'
import { cleanQueryParams } from '@/utils/query'

const authStore = useAuthStore()
const { can } = usePermission()
const { withPlatformViewScope } = usePlatformViewScope()
const isPlatformAdmin = computed(() => authStore.currentUser?.roleCode === 'platform_admin')
const canCreateCourse = computed(() => can({ permission: 'course:create', menuPermission: 'menu:course:view' }))
const canEditCourse = computed(() => can({ permission: 'course:update', menuPermission: 'menu:course:view' }))
const canDeleteCourse = computed(() => can({ permission: 'course:delete', menuPermission: 'menu:course:view' }))
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const rows = ref<CourseItem[]>([])
const total = ref(0)
const query = reactive<CourseSearchParams>({
  pageNum: 1,
  pageSize: 10,
})

const emptyForm: CourseFormModel = {
  courseName: '',
  subject: '',
  courseType: '',
  durationMinutes: 60,
  feeStandard: 0,
  status: 'enabled',
  remark: '',
}

const form = reactive<CourseFormModel>({ ...emptyForm })

onMounted(() => {
  load()
})

async function load(extra?: Partial<CourseSearchParams>) {
  loading.value = true
  try {
    const params = withPlatformViewScope({
      ...query,
      ...extra,
    })
    const response = await fetchCourseList(cleanQueryParams(params))
    const normalized = normalizeListPayload<CourseItem>(response.data?.data)
    rows.value = normalized.list
    total.value = normalized.total
    Object.assign(query, params)
  } finally {
    loading.value = false
  }
}

function handlePageChange(pageNum: number) {
  load({ pageNum })
}

function handleSizeChange(pageSize: number) {
  load({ pageNum: 1, pageSize })
}

function openCreateDialog() {
  if (!canCreateCourse.value) {
    message.warning('没有新增课程权限')
    return
  }
  editingId.value = null
  Object.assign(form, emptyForm)
  dialogVisible.value = true
}

function openEditDialog(row: CourseItem) {
  if (!canEditCourse.value) {
    message.warning('没有编辑课程权限')
    return
  }
  editingId.value = row.id
  Object.assign(form, {
    courseName: row.courseName,
    subject: row.subject,
    courseType: row.courseType,
    durationMinutes: row.durationMinutes,
    feeStandard: row.feeStandard,
    status: row.status,
    remark: row.remark || '',
  })
  dialogVisible.value = true
}

function formatStatus(value: string) {
  if (value === 'enabled') return '启用'
  if (value === 'disabled') return '禁用'
  return value || '-'
}

async function handleDelete(row: CourseItem) {
  if (!canDeleteCourse.value) {
    message.warning('没有删除课程权限')
    return
  }
  try {
    await confirmDialog(`确认删除课程「${row.courseName}」吗？`, '删除确认', { type: 'warning' })
    await deleteCourse(row.id)
    message.success('课程删除成功')
    await load({ pageNum: 1 })
  } catch (error) {
    if (!isCancelError(error)) {
      console.error(error)
      message.error(extractErrorMessage(error, '课程删除失败，请稍后重试'))
    }
  }
}

async function handleSubmit() {
  if (editingId.value && !canEditCourse.value) {
    message.warning('没有编辑课程权限')
    return
  }
  if (!editingId.value && !canCreateCourse.value) {
    message.warning('没有新增课程权限')
    return
  }
  if (!form.courseName.trim() || !form.subject.trim() || !form.courseType.trim()) {
    message.warning('请先填写课程名称、科目和课程类型信息')
    return
  }
  submitting.value = true
  try {
    if (editingId.value) {
      await updateCourse(editingId.value, form)
      message.success('课程更新成功')
    } else {
      await createCourse(form)
      message.success('课程创建成功')
    }
    dialogVisible.value = false
    await load()
  } catch (error) {
    console.error(error)
    message.error(extractErrorMessage(error, '课程保存失败，请检查输入内容'))
  } finally {
    submitting.value = false
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

.table-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
