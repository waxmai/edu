<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">{{ authStore.isTeacher ? '我的学生' : '业务中台' }}</div>
        <h2>{{ authStore.isTeacher ? '我的学生' : '学员管理' }}</h2>
        <p>{{ authStore.isTeacher ? '查看与本人授课相关的学员档案，并快速进入排课与课堂记录。' : '统一管理学员档案，并快速进入排课、收费和详情查看。' }}</p>
      </div>
      <el-button v-if="canCreateStudent" type="primary" @click="openCreateDialog">新增学员</el-button>
    </div>

    <el-card shadow="never" class="toolbar-card search-form-card">
      <StudentSearchForm :model="query" @search="handleSearch" @reset="handleReset" />
    </el-card>

    <el-alert
      :title="pageAlertText"
      type="success"
      :closable="false"
    />

    <el-card shadow="never" class="section-card data-table-card">
      <el-table :data="data" v-loading="loading" border>
        <el-table-column prop="studentName" label="学员姓名" min-width="120" />
        <el-table-column prop="grade" label="年级" min-width="100" />
        <el-table-column prop="subject" label="科目" min-width="100" />
        <el-table-column prop="teachingType" label="授课形式" min-width="120">
          <template #default="scope">{{ formatTeachingType(scope.row.teachingType) }}</template>
        </el-table-column>
        <el-table-column v-if="!authStore.isTeacher" prop="parentName" label="家长姓名" min-width="120" />
        <el-table-column v-if="!authStore.isTeacher" prop="parentPhone" label="家长电话" min-width="140" />
        <el-table-column label="状态" min-width="100">
          <template #default="scope">
            <StudentStatusTag :status="scope.row.status" />
          </template>
        </el-table-column>
        <el-table-column :label="authStore.isTeacher ? '工作台操作' : '操作'" fixed="right" :min-width="authStore.isTeacher ? 220 : 280">
          <template #default="scope">
            <el-space wrap>
              <el-button link type="primary" @click="goDetail(scope.row.id)">详情</el-button>
              <el-button v-if="canEditStudent" link @click="openEditDialog(scope.row)">编辑</el-button>
              <el-button v-if="canDeleteStudent" link type="danger" @click="handleDelete(scope.row)">删除</el-button>
              <el-button v-if="canCreateSchedule" link @click="goCreateSchedule(scope.row.id)">{{ authStore.isTeacher ? '查看排课' : '排课' }}</el-button>
              <el-button v-if="canCreatePayment" link @click="goCreatePayment(scope.row.id)">缴费</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && data.length === 0" :description="emptyText" />

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

    <StudentFormDialog
      v-model="dialogVisible"
      :initial-value="formModel"
      @submit="handleSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createStudent, deleteStudent, updateStudent } from '@/api/student'
import StudentFormDialog from '@/modules/student/components/StudentFormDialog.vue'
import StudentSearchForm from '@/modules/student/components/StudentSearchForm.vue'
import StudentStatusTag from '@/modules/student/components/StudentStatusTag.vue'
import { useStudentList } from '@/modules/student/composables/useStudentList'
import type { StudentFormModel, StudentItem } from '@/modules/student/types/student'
import { useAuthStore } from '@/stores/auth'
import { usePermission } from '@/composables/usePermission'
import { confirmDialog } from '@/utils/confirm'
import { extractErrorMessage, isCancelError } from '@/utils/error'
import { message } from '@/utils/message'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { can } = usePermission()
const { loading, data, total, query, load, reset } = useStudentList()
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const canCreateStudent = computed(() => can({ permission: 'student:create', menuPermission: 'menu:student:view' }))
const canEditStudent = computed(() => can({ permission: 'student:update', menuPermission: 'menu:student:view' }))
const canDeleteStudent = computed(() => can({ permission: 'student:delete', menuPermission: 'menu:student:view' }))
const canCreatePayment = computed(() => can({ permission: 'payment:create', menuPermission: 'menu:payment:view' }))
const canCreateSchedule = computed(() => can({ permission: 'schedule:create', menuPermission: 'menu:schedule:view' }))
const pageAlertText = computed(() => {
  if (authStore.isTeacher) return '当前页面仅展示与本人授课相关的学员，可快速进入个人排课与课堂记录。'
  if (authStore.isAdmin) return '可统一管理学员档案，并快捷进入排课、缴费和详情查看。'
  return '当前账号仅可查看本人权限范围内的学员档案，并进入相关排课与详情。'
})
const emptyText = computed(() => {
  if (authStore.isTeacher) return '当前还没有分配到与您授课相关的学员'
  return authStore.isAdmin ? '还没有学员，可先创建第一位学员' : '当前权限范围内还没有学员数据'
})

const emptyForm: StudentFormModel = {
  studentName: '',
  gender: 'male',
  grade: '',
  phone: '',
  parentName: '',
  parentPhone: '',
  subject: '',
  teachingType: 'one_to_one',
  status: 'active',
  remark: '',
}

const formModel = reactive<StudentFormModel>({ ...emptyForm })

onMounted(() => {
  hydrateFromRoute()
  load()
})

function hydrateFromRoute() {
  if (typeof route.query.status === 'string' && route.query.status.trim()) {
    query.status = route.query.status.trim()
  }
  if (route.query.inactiveAlert === '1') {
    query.inactiveAlert = true
    query.status = query.status || 'active'
  }
}

function handleSearch() {
  load({ pageNum: 1 })
}

function handleReset() {
  reset()
  load({ pageNum: 1 })
}

function handlePageChange(pageNum: number) {
  load({ pageNum })
}

function handleSizeChange(pageSize: number) {
  load({ pageSize, pageNum: 1 })
}

function openCreateDialog() {
  if (!canCreateStudent.value) {
    message.warning('当前角色不能新增学员')
    return
  }
  editingId.value = null
  Object.assign(formModel, emptyForm)
  dialogVisible.value = true
}

function openEditDialog(row: StudentItem) {
  if (!canEditStudent.value) {
    message.warning('当前角色不能编辑学员')
    return
  }
  editingId.value = row.id
  Object.assign(formModel, {
    ...emptyForm,
    studentName: row.studentName,
    gender: row.gender || emptyForm.gender,
    grade: row.grade,
    phone: row.phone || '',
    parentName: row.parentName,
    parentPhone: row.parentPhone,
    subject: row.subject,
    teachingType: row.teachingType,
    status: row.status,
    remark: row.remark || '',
  })
  dialogVisible.value = true
}

function formatTeachingType(value: string) {
  if (value === 'one_to_one') return '一对一'
  if (value === 'one_to_many') return '一对多'
  if (value === 'small_class') return '小班'
  return value || '-'
}

async function handleSubmit(value: StudentFormModel) {
  if (!value.studentName.trim() || !value.grade.trim() || !value.parentName.trim() || !value.parentPhone.trim() || !value.subject.trim()) {
    message.warning('请先填写学员姓名、年级、家长姓名、家长电话和报读科目信息')
    return
  }
  try {
    if (editingId.value) {
      await updateStudent(editingId.value, value)
      message.success('学员信息更新成功')
    } else {
      await createStudent(value)
      message.success('学员新增成功')
    }
    dialogVisible.value = false
    await load()
  } catch (error) {
    console.error(error)
    message.error(extractErrorMessage(error, '保存失败，请检查学员姓名、手机号和家长信息后重试'))
  }
}

async function handleDelete(row: StudentItem) {
  if (!canDeleteStudent.value) {
    message.warning('当前角色不能删除学员')
    return
  }
  try {
    await confirmDialog(`确认删除学员「${row.studentName}」吗？`, '删除确认', { type: 'warning' })
    await deleteStudent(row.id)
    message.success('学员删除成功')
    await load({ pageNum: 1 })
  } catch (error) {
    if (!isCancelError(error)) {
      console.error(error)
      message.error(extractErrorMessage(error, '学员删除失败，请稍后再试'))
    }
  }
}

function goDetail(id: number) {
  router.push(authStore.isTeacher ? `/my-students/${id}` : `/students/${id}`)
}

function goCreateSchedule(id: number) {
  router.push(`/schedules?studentId=${id}&from=${authStore.isTeacher ? 'my-student' : 'student'}`)
}

function goCreatePayment(id: number) {
  router.push(`/payments?studentId=${id}&from=student`)
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
