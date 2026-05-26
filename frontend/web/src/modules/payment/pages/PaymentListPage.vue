<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">{{ pageKicker }}</div>
        <h2>{{ pageTitle }}</h2>
        <p>{{ pageDesc }}</p>
      </div>
      <el-button v-if="canCreatePayment" type="primary" @click="openCreateDialog">新增缴费记录</el-button>
    </div>

    <el-card shadow="never" class="toolbar-card search-form-card">
      <PaymentSearchForm :model="query" @search="handleSearch" @reset="handleReset" />
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
        <el-table-column label="课时包" min-width="180">
          <template #default="scope">{{ scope.row.lessonPackageName || `未命名课时包（#${scope.row.lessonPackageId}）` }}</template>
        </el-table-column>
        <el-table-column prop="paymentType" label="缴费类型" min-width="120">
          <template #default="scope">{{ formatPaymentType(scope.row.paymentType) }}</template>
        </el-table-column>
        <el-table-column prop="amount" label="缴费金额" min-width="120">
          <template #default="scope">¥ {{ Number(scope.row.amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="paymentMethod" label="支付方式" min-width="120">
          <template #default="scope">{{ formatPaymentMethod(scope.row.paymentMethod) }}</template>
        </el-table-column>
        <el-table-column prop="paymentTime" label="缴费时间" min-width="180" />
        <el-table-column label="缴费状态" min-width="120">
          <template #default="scope">
            <PaymentStatusTag :status="scope.row.paymentStatus" />
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
        <el-table-column v-if="canManagePayments" label="操作" fixed="right" min-width="160">
          <template #default="scope">
            <el-space>
              <el-button v-if="canEditPayment" link type="primary" @click="openEditDialog(scope.row)">编辑</el-button>
              <el-button v-if="canDeletePayment" link type="danger" @click="handleDelete(scope.row)">删除</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && data.length === 0" :description="authStore.isAdmin ? '还没有缴费记录，可先登记第一笔收费' : '当前权限范围内还没有缴费记录'" />

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

    <PaymentFormDialog
      v-model="dialogVisible"
      :initial-value="formModel"
      @submit="handleSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { createPayment, deletePayment, updatePayment } from '@/api/payment'
import PaymentFormDialog from '@/modules/payment/components/PaymentFormDialog.vue'
import PaymentSearchForm from '@/modules/payment/components/PaymentSearchForm.vue'
import PaymentStatusTag from '@/modules/payment/components/PaymentStatusTag.vue'
import { paymentMethodOptions, paymentStatusOptions, paymentTypeOptions } from '@/modules/payment/constants/payment'
import { usePaymentList } from '@/modules/payment/composables/usePaymentList'
import type { PaymentFormModel, PaymentItem } from '@/modules/payment/types/payment'
import { useAuthStore } from '@/stores/auth'
import { usePermission } from '@/composables/usePermission'
import { confirmDialog } from '@/utils/confirm'
import { extractErrorMessage, isCancelError } from '@/utils/error'
import { normalizeEnumValue } from '@/utils/enum'
import { message } from '@/utils/message'

const route = useRoute()
const authStore = useAuthStore()
const { can } = usePermission()
const { loading, data, total, query, load, reset } = usePaymentList()
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const isOrgAdmin = computed(() => authStore.currentUser?.roleCode === 'org_admin')
const isCampusAdmin = computed(() => authStore.currentUser?.roleCode === 'campus_admin')
const pageKicker = computed(() => (isOrgAdmin.value ? '机构经营收款' : isCampusAdmin.value ? '校区现场收款' : '财务视图'))
const pageTitle = computed(() => (isOrgAdmin.value ? '机构收费管理' : isCampusAdmin.value ? '校区收费管理' : '收费管理'))
const pageDesc = computed(() => (isOrgAdmin.value ? '统一查看本机构各校区缴费情况，关注续费、收款与经营回款。' : isCampusAdmin.value ? '登记和跟进本校区学员缴费，处理现场续费与到款记录。' : '统一查看缴费状态、支付方式和课时包收费进度。'))
const pageAlert = computed(() => (isOrgAdmin.value ? '当前页面面向机构经营与财务协同，可按学员、课时包、状态和支付方式统一筛选。' : isCampusAdmin.value ? '当前页面聚焦本校区收费登记与跟进，支持现场新增、修改与删除收费记录。' : authStore.isAdmin ? '可统一查看和维护缴费记录，并按学员、课时包、缴费状态和支付方式筛选。' : '当前账号可查看本人权限范围内的缴费记录，新增与修改仅管理员可执行。'))
const canCreatePayment = computed(() => can({ permission: 'payment:create', menuPermission: 'menu:payment:view' }))
const canEditPayment = computed(() => can({ permission: 'payment:update', menuPermission: 'menu:payment:view' }))
const canDeletePayment = computed(() => can({ permission: 'payment:delete', menuPermission: 'menu:payment:view' }))
const canManagePayments = computed(() => canCreatePayment.value || canEditPayment.value || canDeletePayment.value)

const emptyForm: PaymentFormModel = {
  studentId: null,
  lessonPackageId: null,
  paymentType: 'signup',
  amount: 0,
  paymentMethod: 'wechat',
  paymentTime: '',
  paymentStatus: 'paid',
  remark: '',
}

const formModel = reactive<PaymentFormModel>({ ...emptyForm })

onMounted(() => {
  hydrateFromRoute()
  load({ ...query })
  if (route.query.from && canCreatePayment.value) {
    openCreateDialog()
  }
})

function hydrateFromRoute() {
  if (typeof route.query.studentId === 'string' && /^\d+$/.test(route.query.studentId.trim())) {
    query.studentId = route.query.studentId.trim()
    formModel.studentId = Number(route.query.studentId) || null
  }
  if (typeof route.query.lessonPackageId === 'string' && /^\d+$/.test(route.query.lessonPackageId.trim())) {
    query.lessonPackageId = route.query.lessonPackageId.trim()
    formModel.lessonPackageId = Number(route.query.lessonPackageId) || null
  }
  const paymentTypeValues = paymentTypeOptions.map((item) => item.value).filter(Boolean) as string[]
  const normalizedPaymentType = normalizeEnumValue(route.query.paymentType, paymentTypeValues)
  if (normalizedPaymentType) {
    query.paymentType = normalizedPaymentType
    formModel.paymentType = normalizedPaymentType
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
  if (!canCreatePayment.value) {
    message.warning('当前角色不能新增缴费记录')
    return
  }
  editingId.value = null
  const preservedStudentId = formModel.studentId
  const preservedLessonPackageId = formModel.lessonPackageId
  const preservedPaymentType = formModel.paymentType
  Object.assign(formModel, emptyForm)
  formModel.studentId = preservedStudentId
  formModel.lessonPackageId = preservedLessonPackageId
  formModel.paymentType = preservedPaymentType
  dialogVisible.value = true
}

function openEditDialog(row: PaymentItem) {
  if (!canEditPayment.value) {
    message.warning('当前角色不能编辑缴费记录')
    return
  }
  editingId.value = row.id
  Object.assign(formModel, {
    studentId: row.studentId,
    lessonPackageId: row.lessonPackageId,
    paymentType: row.paymentType,
    amount: row.amount,
    paymentMethod: row.paymentMethod,
    paymentTime: normalizePaymentTime(row.paymentTime),
    paymentStatus: row.paymentStatus,
    remark: row.remark || '',
  })
  dialogVisible.value = true
}

function formatPaymentType(value: string) {
  return value === 'renewal' ? '续费缴费' : value === 'signup' ? '报名缴费' : value || '-'
}

function formatPaymentMethod(value: string) {
  switch (value) {
    case 'cash':
      return '现金'
    case 'wechat':
      return '微信'
    case 'alipay':
      return '支付宝'
    case 'bank_transfer':
      return '银行转账'
    case 'other':
      return '其他'
    default:
      return value || '-'
  }
}

function normalizePaymentTime(value?: string) {
  if (!value) {
    return ''
  }
  return value.replace('T', ' ').slice(0, 19)
}

async function handleDelete(row: PaymentItem) {
  if (!canDeletePayment.value) {
    message.warning('当前角色只有查看权限，如需删除缴费记录，请使用机构管理员或校区管理员账号')
    return
  }
  try {
    await confirmDialog(`确认删除${row.studentName ? `「${row.studentName}」的` : ''}缴费记录吗？`, '删除确认', { type: 'warning' })
    await deletePayment(row.id)
    message.success('缴费记录删除成功')
    await load({ pageNum: 1 })
  } catch (error) {
    if (!isCancelError(error)) {
      console.error(error)
      message.error(extractErrorMessage(error, '缴费记录删除失败，请稍后再试'))
    }
  }
}

async function handleSubmit(value: PaymentFormModel) {
  if (!canManagePayments.value) {
    message.warning('当前角色只有查看权限，如需新增或修改缴费记录，请使用机构管理员或校区管理员账号')
    return
  }
  if (!value.studentId || !value.lessonPackageId || !value.paymentTime || value.amount <= 0) {
    message.warning('请先选择学员、课时包，并填写缴费时间和金额')
    return
  }
  try {
    if (editingId.value) {
      await updatePayment(editingId.value, value)
      message.success('已更新缴费记录')
    } else {
      await createPayment(value)
      message.success('已新增缴费记录')
    }
    dialogVisible.value = false
    await load()
  } catch (error) {
    console.error(error)
    const rawMessage = extractErrorMessage(error, '保存失败，请检查学员、课时包、缴费时间和金额后重试')
    if (rawMessage.includes('已缴金额不能超过总金额')) {
      message.error('保存失败，当前缴费会导致课时包的已缴金额超过总金额，请先调整课时包金额或本次缴费金额')
      return
    }
    message.error(rawMessage)
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
