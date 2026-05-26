<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">教师工作台</div>
        <h2>我的上课记录</h2>
        <p>查看并维护与本人授课相关的课堂记录、作业与反馈。</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">新增上课记录</el-button>
    </div>

    <el-alert
      type="success"
      :closable="false"
      title="当前页面仅展示与本人授课相关的上课记录，可维护课堂内容、作业与反馈。"
    />

    <el-card shadow="never" class="toolbar-card search-form-card">
      <el-form :inline="true" :model="query">
        <el-form-item label="学员">
          <el-select v-model="query.studentId" clearable filterable placeholder="请选择学员" style="width: 200px">
            <el-option v-for="item in studentOptions" :key="item.id" :label="`${item.studentName}（#${item.id}）`" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="出勤状态">
          <el-select v-model="query.attendanceStatus" clearable style="width: 160px">
            <el-option label="到课" value="present" />
            <el-option label="迟到" value="late" />
            <el-option label="缺勤" value="absent" />
            <el-option label="请假" value="leave" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="section-card data-table-card" v-loading="loading">
      <el-table :data="rows" border>
        <el-table-column label="对应排课" min-width="180">
          <template #default="scope">{{ formatScheduleLabel(scope.row) }}</template>
        </el-table-column>
        <el-table-column label="学员" min-width="160">
          <template #default="scope">{{ scope.row.studentName || (scope.row.studentId ? `未命名学员（#${scope.row.studentId}）` : '-') }}</template>
        </el-table-column>
        <el-table-column prop="attendanceStatus" label="出勤" min-width="100">
          <template #default="scope">{{ formatAttendanceStatus(scope.row.attendanceStatus) }}</template>
        </el-table-column>
        <el-table-column prop="deductLessonCount" label="扣减课时" min-width="100" />
        <el-table-column prop="lessonDeducted" label="已扣课时" min-width="100">
          <template #default="scope">{{ scope.row.lessonDeducted ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column prop="lessonContent" label="上课内容" min-width="180" show-overflow-tooltip />
        <el-table-column prop="homework" label="作业" min-width="180" show-overflow-tooltip />
        <el-table-column prop="feedback" label="反馈" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" fixed="right" min-width="160">
          <template #default="scope">
            <el-space>
              <el-button link type="primary" @click="openEditDialog(scope.row)">编辑</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && rows.length === 0" description="还没有上课记录，可先登记第一条课堂记录" />

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

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑上课记录' : '新增上课记录'" width="720px">
      <el-form :model="form" label-width="110px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="排课">
              <el-select v-model="form.scheduleId" clearable filterable placeholder="请选择排课" style="width: 100%" @change="handleScheduleChange">
                <el-option v-for="item in scheduleOptions" :key="item.id" :label="formatScheduleOption(item)" :value="item.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12"><el-form-item label="学员"><el-input-number v-model="form.studentId" :min="1" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="教师"><el-input-number v-model="form.teacherId" :min="1" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="出勤状态"><el-select v-model="form.attendanceStatus" style="width: 100%"><el-option label="到课" value="present" /><el-option label="迟到" value="late" /><el-option label="缺勤" value="absent" /><el-option label="请假" value="leave" /></el-select></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="上课内容"><el-input v-model="form.lessonContent" type="textarea" :rows="3" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="作业"><el-input v-model="form.homework" type="textarea" :rows="3" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="反馈"><el-input v-model="form.feedback" type="textarea" :rows="3" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="扣课时"><el-switch v-model="form.needDeductLesson" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="扣减课时数"><el-input-number v-model="form.deductLessonCount" :min="0" :step="1" style="width: 100%" /></el-form-item></el-col>
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
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  createLessonRecord,
  fetchLessonRecordList,
  updateLessonRecord,
  type LessonRecordFormModel,
  type LessonRecordItem,
  type LessonRecordSearchParams,
} from '@/api/lessonRecord'
import { useFormOptions } from '@/composables/useFormOptions'
import { useAuthStore } from '@/stores/auth'
import { normalizeListPayload } from '@/utils/api'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'
import { cleanQueryParams } from '@/utils/query'

const route = useRoute()
const authStore = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const rows = ref<LessonRecordItem[]>([])
const total = ref(0)
const query = reactive<LessonRecordSearchParams>({
  studentId: '',
  teacherId: '',
  attendanceStatus: '',
  pageNum: 1,
  pageSize: 10,
})
const { studentOptions, scheduleOptions, loadStudentOptions, loadScheduleOptions } = useFormOptions()

const emptyForm: LessonRecordFormModel = {
  scheduleId: null,
  studentId: null,
  teacherId: null,
  attendanceStatus: 'present',
  lessonContent: '',
  homework: '',
  feedback: '',
  needDeductLesson: true,
  deductLessonCount: 1,
}

const form = reactive<LessonRecordFormModel>({ ...emptyForm })

onMounted(async () => {
  await loadStudentOptions()
  if (authStore.currentUser?.id) {
    query.teacherId = String(authStore.currentUser.id)
    form.teacherId = authStore.currentUser.id
  }
  await loadScheduleOptions({ teacherId: query.teacherId })
  if (typeof route.query.studentId === 'string' && /^\d+$/.test(route.query.studentId.trim())) {
    query.studentId = route.query.studentId.trim()
    form.studentId = Number(route.query.studentId) || null
    await loadScheduleOptions({ studentId: query.studentId, teacherId: query.teacherId })
  }
  loadRecords()
})

async function loadRecords(extra?: Partial<LessonRecordSearchParams>) {
  loading.value = true
  try {
    const params = { ...query, ...extra }
    const response = await fetchLessonRecordList(cleanQueryParams(params))
    const normalized = normalizeListPayload<LessonRecordItem>(response.data?.data)
    rows.value = normalized.list
    total.value = normalized.total
    Object.assign(query, params)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  loadScheduleOptions({ studentId: query.studentId, teacherId: query.teacherId })
  loadRecords({ pageNum: 1 })
}

function formatAttendanceStatus(value: string) {
  switch (value) {
    case 'present':
      return '到课'
    case 'late':
      return '迟到'
    case 'absent':
      return '缺勤'
    case 'leave':
      return '请假'
    default:
      return value || '-'
  }
}

function resetQuery() {
  query.studentId = ''
  query.teacherId = authStore.currentUser?.id ? String(authStore.currentUser.id) : ''
  query.attendanceStatus = ''
  query.pageNum = 1
  loadScheduleOptions({ teacherId: query.teacherId })
  loadRecords({ pageNum: 1 })
}

function handlePageChange(pageNum: number) {
  loadRecords({ pageNum })
}

function handleSizeChange(pageSize: number) {
  loadRecords({ pageSize, pageNum: 1 })
}

function openCreateDialog() {
  const preservedStudentId = form.studentId
  Object.assign(form, { ...emptyForm })
  form.studentId = preservedStudentId
  form.teacherId = authStore.currentUser?.id || null
  loadScheduleOptions({
    studentId: form.studentId || undefined,
    teacherId: form.teacherId || undefined,
  })
  editingId.value = null
  dialogVisible.value = true
}

function openEditDialog(row: LessonRecordItem) {
  editingId.value = row.id
  Object.assign(form, {
    scheduleId: row.scheduleId,
    studentId: row.studentId,
    teacherId: row.teacherId,
    attendanceStatus: row.attendanceStatus,
    lessonContent: row.lessonContent || '',
    homework: row.homework || '',
    feedback: row.feedback || '',
    needDeductLesson: row.lessonDeducted,
    deductLessonCount: row.deductLessonCount ?? 1,
  })
  dialogVisible.value = true
}

function handleScheduleChange(scheduleId?: number) {
  const matched = scheduleOptions.value.find((item) => item.id === scheduleId)
  if (!matched) return
  form.studentId = matched.studentId || form.studentId
  form.teacherId = matched.teacherId || form.teacherId
}

function formatScheduleOption(item: any) {
  return [item.classDate || '未排日期', item.startTime || '', item.studentName || '', item.courseName || ''].filter(Boolean).join(' · ')
}

function formatScheduleLabel(item: any) {
  return [item.classDate || '', item.startTime || '', item.courseName || ''].filter(Boolean).join(' · ') || `排课 #${item.scheduleId}`
}

async function handleSubmit() {
  if (!form.scheduleId || !form.studentId || !form.teacherId) {
    message.warning('请先选择排课、学员和教师')
    return
  }
  submitting.value = true
  try {
    if (editingId.value) {
      await updateLessonRecord(editingId.value, form)
      message.success('上课记录更新成功')
    } else {
      await createLessonRecord(form)
      message.success('上课记录新增成功')
    }
    dialogVisible.value = false
    loadRecords({ pageNum: 1 })
  } catch (error) {
    message.error(extractErrorMessage(error, '保存上课记录失败，请稍后重试'))
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
