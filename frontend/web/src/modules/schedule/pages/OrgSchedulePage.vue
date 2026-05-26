<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">机构排课统筹</div>
        <h2>机构排课管理</h2>
        <p>统筹本机构排课安排、资源协调与调补课衔接。</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">新增排课</el-button>
    </div>

    <el-card shadow="never" class="toolbar-card search-form-card">
      <el-form :inline="true" :model="query" @submit.prevent>
        <el-form-item label="学员">
          <el-select v-model="query.studentId" clearable filterable placeholder="请选择学员" style="width: 200px">
            <el-option v-for="item in studentOptions" :key="item.id" :label="`${item.studentName}（#${item.id}）`" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="教师">
          <el-select v-model="query.teacherId" clearable filterable placeholder="请选择教师" style="width: 200px">
            <el-option v-for="item in teacherOptions" :key="item.id" :label="`${item.realName}（${item.username}）`" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="课时包">
          <el-select v-model="query.lessonPackageId" clearable filterable placeholder="请选择课时包" style="width: 220px">
            <el-option v-for="item in lessonPackageOptions" :key="item.id" :label="lessonPackageOptionLabel(item)" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.scheduleStatus" clearable style="width: 160px">
            <el-option label="待上课" value="scheduled" />
            <el-option label="已完成" value="completed" />
            <el-option label="已请假" value="leave" />
            <el-option label="已调课" value="rescheduled" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="待补课" value="makeup_pending" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="section-card data-table-card">
      <el-table :data="data" v-loading="loading" border>
        <el-table-column label="学员" min-width="160">
          <template #default="scope">{{ scope.row.studentName || (scope.row.studentId ? `未命名学员（#${scope.row.studentId}）` : '-') }}</template>
        </el-table-column>
        <el-table-column label="课程" min-width="160">
          <template #default="scope">{{ scope.row.courseName || (scope.row.courseId ? `未命名课程（#${scope.row.courseId}）` : '-') }}</template>
        </el-table-column>
        <el-table-column label="教师" min-width="160">
          <template #default="scope">{{ scope.row.teacherName || (scope.row.teacherId ? `未命名教师（#${scope.row.teacherId}）` : '-') }}</template>
        </el-table-column>
        <el-table-column prop="classDate" label="上课日期" min-width="120" />
        <el-table-column prop="startTime" label="开始时间" min-width="160" />
        <el-table-column prop="endTime" label="结束时间" min-width="160" />
        <el-table-column prop="classroom" label="教室" min-width="120" />
        <el-table-column label="状态" min-width="120">
          <template #default="scope">
            <ScheduleStatusTag :status="scope.row.scheduleStatus || scope.row.status" />
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" fixed="right" min-width="380">
          <template #default="scope">
            <el-space wrap>
              <el-button link type="primary" @click="openDetail(scope.row)">详情</el-button>
              <el-button link @click="openReschedule(scope.row)">调课</el-button>
              <el-button link type="success" @click="openMakeupDialog(scope.row)">补课</el-button>
              <el-button link type="warning" @click="openLeaveDialog(scope.row)">请假</el-button>
              <el-button link type="danger" @click="openCancelDialog(scope.row)">取消</el-button>
              <el-button link type="danger" @click="handleDelete(scope.row)">删除</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && data.length === 0" description="还没有排课安排，可先创建第一条排课" />

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

    <ScheduleFormDialog
      v-model="dialogVisible"
      title="新增排课"
      :initial-value="form"
      @submit="handleCreate"
    />

    <ScheduleDetailDrawer v-model="detailVisible" :schedule="selectedSchedule" />

    <el-dialog v-model="reasonDialogVisible" :title="reasonDialogMode === 'cancel' ? '取消排课' : '学员请假'" width="520px">
      <el-form label-position="top">
        <el-form-item :label="reasonDialogMode === 'cancel' ? '取消原因' : '请假原因'">
          <el-input v-model="reasonForm.reason" type="textarea" :rows="4" placeholder="请输入原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reasonDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="reasonSubmitting" @click="submitReasonAction">确认</el-button>
      </template>
    </el-dialog>

    <ScheduleFormDialog
      v-model="makeupDialogVisible"
      title="新增补课排课"
      :initial-value="makeupForm"
      @submit="handleMakeupCreate"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { cancelSchedule, createMakeupSchedule, createSchedule, deleteSchedule, leaveSchedule } from '@/api/schedule'
import { useFormOptions } from '@/composables/useFormOptions'
import ScheduleDetailDrawer from '@/modules/schedule/components/ScheduleDetailDrawer.vue'
import ScheduleFormDialog from '@/modules/schedule/components/ScheduleFormDialog.vue'
import ScheduleStatusTag from '@/modules/schedule/components/ScheduleStatusTag.vue'
import { useScheduleCalendar } from '@/modules/schedule/composables/useScheduleCalendar'
import type { ScheduleCreatePayload, ScheduleListItem } from '@/api/schedule'
import { confirmDialog } from '@/utils/confirm'
import { extractErrorMessage, isCancelError } from '@/utils/error'
import { message } from '@/utils/message'

const route = useRoute()
const router = useRouter()
const { loading, data, total, query, load, reset } = useScheduleCalendar()
const { studentOptions, teacherOptions, lessonPackageOptions, loadBaseOptions, loadLessonPackageOptions, lessonPackageOptionLabel } = useFormOptions()
const dialogVisible = ref(false)
const detailVisible = ref(false)
const submitting = ref(false)
const selectedSchedule = ref<ScheduleListItem | null>(null)
const reasonDialogVisible = ref(false)
const reasonDialogMode = ref<'cancel' | 'leave'>('cancel')
const reasonSubmitting = ref(false)
const actionScheduleId = ref<number | null>(null)
const makeupDialogVisible = ref(false)
const makeupOriginalScheduleId = ref<number | null>(null)

const defaultForm = {
  studentId: null as number | null,
  courseId: null as number | null,
  teacherId: null as number | null,
  lessonPackageId: null as number | null,
  classDate: '',
  startTime: '',
  endTime: '',
  classroom: '',
  scheduleStatus: 'scheduled',
  remark: '',
}

const form = reactive({ ...defaultForm })
const makeupForm = reactive({ ...defaultForm })
const reasonForm = reactive({ reason: '' })

onMounted(async () => {
  await loadBaseOptions()
  await loadLessonPackageOptions()
  if (typeof route.query.studentId === 'string' && /^\d+$/.test(route.query.studentId.trim()) && Number(route.query.studentId) > 0) {
    form.studentId = Number(route.query.studentId)
    query.studentId = route.query.studentId.trim()
    dialogVisible.value = Boolean(route.query.from)
  }
  if (typeof route.query.lessonPackageId === 'string' && /^\d+$/.test(route.query.lessonPackageId.trim()) && Number(route.query.lessonPackageId) > 0) {
    form.lessonPackageId = Number(route.query.lessonPackageId)
    query.lessonPackageId = route.query.lessonPackageId.trim()
  }
  load({ ...query })
})

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
  Object.assign(form, defaultForm)
  if (typeof route.query.studentId === 'string' && /^\d+$/.test(route.query.studentId.trim()) && Number(route.query.studentId) > 0) {
    form.studentId = Number(route.query.studentId)
  }
  if (typeof route.query.lessonPackageId === 'string' && /^\d+$/.test(route.query.lessonPackageId.trim()) && Number(route.query.lessonPackageId) > 0) {
    form.lessonPackageId = Number(route.query.lessonPackageId)
  }
  dialogVisible.value = true
}

function openDetail(row: ScheduleListItem) {
  selectedSchedule.value = row
  detailVisible.value = true
}

function openReschedule(row: ScheduleListItem) {
  router.push(`/reschedules?scheduleId=${row.id}`)
}

function openMakeupDialog(row: ScheduleListItem) {
  makeupOriginalScheduleId.value = row.id
  Object.assign(makeupForm, {
    ...defaultForm,
    studentId: row.studentId || null,
    courseId: row.courseId || null,
    teacherId: row.teacherId || null,
    lessonPackageId: row.lessonPackageId || null,
    classroom: row.classroom || '',
    scheduleStatus: 'scheduled',
  })
  makeupDialogVisible.value = true
}

function openLeaveDialog(row: ScheduleListItem) {
  reasonDialogMode.value = 'leave'
  actionScheduleId.value = row.id
  reasonForm.reason = ''
  reasonDialogVisible.value = true
}

function openCancelDialog(row: ScheduleListItem) {
  reasonDialogMode.value = 'cancel'
  actionScheduleId.value = row.id
  reasonForm.reason = ''
  reasonDialogVisible.value = true
}

async function handleDelete(row: ScheduleListItem) {
  try {
    await confirmDialog(`确认删除 ${row.studentName || `学员 #${row.studentId}`} 的排课安排吗？`, '删除确认', { type: 'warning' })
    await deleteSchedule(row.id)
    message.success('排课删除成功')
    load({ pageNum: 1 })
  } catch (error) {
    if (isCancelError(error)) return
    message.error(extractErrorMessage(error, '删除排课失败，请稍后重试'))
  }
}

async function handleCreate(payload: ScheduleCreatePayload) {
  submitting.value = true
  try {
    await createSchedule(payload)
    message.success('排课创建成功')
    dialogVisible.value = false
    load({ pageNum: 1 })
  } catch (error) {
    message.error(extractErrorMessage(error, '排课创建失败，请检查时间冲突或课时包状态'))
  } finally {
    submitting.value = false
  }
}

async function submitReasonAction() {
  if (!actionScheduleId.value) return
  reasonSubmitting.value = true
  try {
    if (reasonDialogMode.value === 'cancel') {
      await cancelSchedule(actionScheduleId.value, { reason: reasonForm.reason })
      message.success('排课已取消')
    } else {
      await leaveSchedule(actionScheduleId.value, { reason: reasonForm.reason })
      message.success('请假已登记')
    }
    reasonDialogVisible.value = false
    load({ pageNum: 1 })
  } catch (error) {
    message.error(extractErrorMessage(error, '操作失败，请稍后重试'))
  } finally {
    reasonSubmitting.value = false
  }
}

async function handleMakeupCreate(payload: ScheduleCreatePayload) {
  if (!makeupOriginalScheduleId.value) {
    message.error('缺少原始排课，无法创建补课')
    return
  }
  submitting.value = true
  try {
    await createMakeupSchedule({
      originalScheduleId: makeupOriginalScheduleId.value,
      studentId: payload.studentId,
      teacherId: payload.teacherId,
      courseId: payload.courseId,
      lessonPackageId: payload.lessonPackageId || null,
      classDate: payload.classDate,
      startTime: payload.startTime,
      endTime: payload.endTime,
      classroom: payload.classroom,
      remark: payload.remark,
    })
    message.success('补课排课创建成功')
    makeupDialogVisible.value = false
    load({ pageNum: 1 })
  } catch (error) {
    message.error(extractErrorMessage(error, '补课创建失败，请检查时间冲突或原始排课状态'))
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
