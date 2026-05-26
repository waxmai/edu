<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">{{ pageKicker }}</div>
        <h2>{{ pageTitle }}</h2>
        <p>{{ pageDesc }}</p>
      </div>
      <el-button v-if="canCreateLessonPackage" type="primary" @click="openCreateDialog">新增课时包</el-button>
    </div>

    <el-card shadow="never" class="toolbar-card search-form-card">
      <LessonPackageSearchForm :model="query" @search="handleSearch" @reset="handleReset" />
    </el-card>

    <el-alert
      :title="pageAlert"
      type="success"
      :closable="false"
    />

    <el-card shadow="never" class="section-card data-table-card">
      <el-table :data="data" v-loading="loading" border>
        <el-table-column prop="studentName" label="学员" min-width="140">
          <template #default="scope">{{ scope.row.studentName || `未命名学员（#${scope.row.studentId}）` }}</template>
        </el-table-column>
        <el-table-column prop="courseName" label="课程" min-width="160">
          <template #default="scope">{{ scope.row.courseName || `未命名课程（#${scope.row.courseId}）` }}</template>
        </el-table-column>
        <el-table-column prop="subject" label="科目" min-width="100" />
        <el-table-column prop="totalLessons" label="总课时" min-width="100" />
        <el-table-column prop="usedLessons" label="已用课时" min-width="100" />
        <el-table-column prop="remainLessons" label="剩余课时" min-width="100" />
        <el-table-column prop="totalAmount" label="总金额" min-width="120">
          <template #default="scope">¥ {{ Number(scope.row.totalAmount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="paidAmount" label="已缴金额" min-width="120">
          <template #default="scope">¥ {{ Number(scope.row.paidAmount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="startDate" label="生效日期" min-width="120" />
        <el-table-column prop="endDate" label="截止日期" min-width="120" />
        <el-table-column label="状态" min-width="100">
          <template #default="scope">
            <LessonPackageStatusTag :status="scope.row.status" />
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" min-width="240">
          <template #default="scope">
            <el-space wrap>
              <el-button v-if="canEditLessonPackage" link type="primary" @click="openEditDialog(scope.row)">编辑</el-button>
              <el-button v-if="canDeleteLessonPackage" link type="danger" @click="handleDelete(scope.row)">删除</el-button>
              <el-button v-if="canCreatePayment" link @click="goRenewPayment(scope.row)">续费</el-button>
              <el-button v-if="canCreatePayment" link @click="goSignupPayment(scope.row)">缴费</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && data.length === 0" :description="authStore.isAdmin ? '还没有课时包，可先为学员创建第一份课时包' : '当前权限范围内还没有可查看的课时包'" />

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

    <LessonPackageFormDialog
      v-model="dialogVisible"
      :initial-value="formModel"
      @submit="handleSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useRouter } from 'vue-router'
import { createLessonPackage, deleteLessonPackage, updateLessonPackage, type LessonPackageUpdatePayload } from '@/api/lessonPackage'
import LessonPackageFormDialog from '@/modules/lesson-package/components/LessonPackageFormDialog.vue'
import LessonPackageSearchForm from '@/modules/lesson-package/components/LessonPackageSearchForm.vue'
import LessonPackageStatusTag from '@/modules/lesson-package/components/LessonPackageStatusTag.vue'
import { useLessonPackageList } from '@/modules/lesson-package/composables/useLessonPackageList'
import type { LessonPackageFormModel, LessonPackageItem } from '@/modules/lesson-package/types/lesson-package'
import { useAuthStore } from '@/stores/auth'
import { usePermission } from '@/composables/usePermission'
import { confirmDialog } from '@/utils/confirm'
import { extractErrorMessage, isCancelError } from '@/utils/error'
import { message } from '@/utils/message'

interface LessonPackageCreatePayload extends LessonPackageFormModel {
  usedLessons: number
  remainLessons: number
}

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { can } = usePermission()
const { loading, data, total, query, load, reset } = useLessonPackageList()
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const isOrgAdmin = computed(() => authStore.currentUser?.roleCode === 'org_admin')
const isCampusAdmin = computed(() => authStore.currentUser?.roleCode === 'campus_admin')
const pageKicker = computed(() => (isOrgAdmin.value ? '机构课时经营' : isCampusAdmin.value ? '校区课时经营' : '课时经营'))
const pageTitle = computed(() => (isOrgAdmin.value ? '机构课时包管理' : isCampusAdmin.value ? '校区课时包管理' : '课时包管理'))
const pageDesc = computed(() => (isOrgAdmin.value ? '统一掌握本机构学员课时余额、有效期与收款进度。' : isCampusAdmin.value ? '维护本校区学员课时包，聚焦剩余课时、到期与续费。' : '统一掌握课时余额、有效期、应收与实收情况。'))
const pageAlert = computed(() => (isOrgAdmin.value ? '当前页面面向机构经营管理，可统一查看剩余课时、有效期和缴费进度。' : isCampusAdmin.value ? '当前页面聚焦本校区课时包维护，便于处理续费、缴费和到期跟进。' : authStore.isAdmin ? '可统一管理学员课时包，查看剩余课时、有效期和缴费情况。' : '当前账号可查看本人相关课时包与缴费进度，新增和维护能力按权限开放。'))
const canCreateLessonPackage = computed(() => can({ permission: 'lesson-package:create', menuPermission: 'menu:lesson-package:view' }))
const canEditLessonPackage = computed(() => can({ permission: 'lesson-package:update', menuPermission: 'menu:lesson-package:view' }))
const canDeleteLessonPackage = computed(() => can({ permission: 'lesson-package:delete', menuPermission: 'menu:lesson-package:view' }))
const canCreatePayment = computed(() => can({ permission: 'payment:create', menuPermission: 'menu:payment:view' }))
const canManageLessonPackage = computed(() => canCreateLessonPackage.value || canEditLessonPackage.value || canDeleteLessonPackage.value)

const emptyForm: LessonPackageFormModel = {
  studentId: null,
  courseId: null,
  totalLessons: 0,
  totalAmount: 0,
  paidAmount: 0,
  startDate: '',
  endDate: '',
  status: 'pending',
  lowLessonThreshold: 3,
  remark: '',
}

const formModel = reactive<LessonPackageFormModel>({ ...emptyForm })

onMounted(() => {
  hydrateFromRoute()
  load({ ...query })
})

function hydrateFromRoute() {
  if (typeof route.query.studentId === 'string' && /^\d+$/.test(route.query.studentId.trim())) {
    query.studentId = route.query.studentId.trim()
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
  if (!canCreateLessonPackage.value) {
    message.warning('当前角色不能新增课时包')
    return
  }
  editingId.value = null
  Object.assign(formModel, emptyForm)
  dialogVisible.value = true
}

function openEditDialog(row: LessonPackageItem) {
  if (!canEditLessonPackage.value) {
    message.warning('当前角色不能编辑课时包')
    return
  }
  editingId.value = row.id
  Object.assign(formModel, {
    ...emptyForm,
    studentId: row.studentId,
    courseId: row.courseId,
    totalLessons: row.totalLessons,
    totalAmount: row.totalAmount,
    paidAmount: row.paidAmount,
    startDate: row.startDate || '',
    endDate: row.endDate || '',
    status: row.status,
    lowLessonThreshold: row.lowLessonThreshold || 3,
    remark: row.remark || '',
  })
  dialogVisible.value = true
}

async function handleSubmit(value: LessonPackageFormModel) {
  if (!value.studentId || !value.courseId || value.totalLessons <= 0) {
    message.warning('请先选择学员、课程并填写总课时')
    return
  }
  if (value.paidAmount > value.totalAmount) {
    message.warning('已缴金额不能超过总金额')
    return
  }
  try {
    if (editingId.value) {
      const updatePayload: LessonPackageUpdatePayload = {
        courseId: value.courseId,
        totalLessons: value.totalLessons,
        totalAmount: value.totalAmount,
        paidAmount: value.paidAmount,
        startDate: value.startDate,
        endDate: value.endDate,
        status: value.status,
        lowLessonThreshold: value.lowLessonThreshold,
        remark: value.remark,
      }
      await updateLessonPackage(editingId.value, updatePayload)
      message.success('已更新课时包')
    } else {
      await createLessonPackage({
        ...(value as LessonPackageFormModel),
        usedLessons: 0,
        remainLessons: value.totalLessons,
      } as LessonPackageCreatePayload)
      message.success('已新增课时包')
    }
    dialogVisible.value = false
    await load()
  } catch (error) {
    console.error(error)
    const rawMessage = extractErrorMessage(error, '保存失败，请检查学员、课程、课时和生效日期后重试')
    if (rawMessage.includes('chk_lesson_package_paid_not_exceed_total')) {
      message.error('已缴金额不能超过总金额')
      return
    }
    message.error(rawMessage)
  }
}

async function handleDelete(row: LessonPackageItem) {
  if (!canDeleteLessonPackage.value) {
    message.warning('当前角色不能删除课时包')
    return
  }
  try {
    await confirmDialog(`确认删除${row.studentName ? `「${row.studentName}」的` : ''}课时包吗？`, '删除确认', { type: 'warning' })
    await deleteLessonPackage(row.id)
    message.success('课时包删除成功')
    await load({ pageNum: 1 })
  } catch (error) {
    if (!isCancelError(error)) {
      console.error(error)
      message.error(extractErrorMessage(error, '课时包删除失败，请稍后再试'))
    }
  }
}

function goSignupPayment(row: LessonPackageItem) {
  if (!canManageLessonPackage.value) {
    message.warning('当前角色不能为课时包发起缴费')
    return
  }
  router.push(`/payments?studentId=${row.studentId}&lessonPackageId=${row.id}&paymentType=signup`)
}

function goRenewPayment(row: LessonPackageItem) {
  if (!canManageLessonPackage.value) {
    message.warning('当前角色不能为课时包发起续费')
    return
  }
  router.push(`/payments?studentId=${row.studentId}&lessonPackageId=${row.id}&paymentType=renewal`)
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
