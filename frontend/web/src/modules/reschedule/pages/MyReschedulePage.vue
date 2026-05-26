<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">教师申请流</div>
        <h2>我的调课申请</h2>
        <p>发起并查看与本人授课相关的调课、请假申请。</p>
      </div>
      <el-button type="primary" @click="dialogVisible = true">发起申请</el-button>
    </div>

    <el-alert
      title="当前页面聚焦本人相关申请记录，便于跟踪处理进度。"
      type="success"
      :closable="false"
    />

    <el-card shadow="never" class="toolbar-card search-form-card">
      <el-form :inline="true" :model="query">
        <el-form-item label="学员">
          <el-select v-model="query.studentId" clearable filterable placeholder="请选择学员" style="width: 200px">
            <el-option v-for="item in studentOptions" :key="item.id" :label="formatStudentOption(item)" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="query.operationType" clearable style="width: 160px">
            <el-option label="调课" value="reschedule" />
            <el-option label="请假" value="leave" />
            <el-option label="补课" value="makeup" />
            <el-option label="取消" value="cancel" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="section-card data-table-card" v-loading="loadingList">
      <el-table :data="records" border>
        <el-table-column label="原排课" min-width="220">
          <template #default="scope">{{ scope.row.oldScheduleSummary || formatScheduleRef(scope.row.oldScheduleId) }}</template>
        </el-table-column>
        <el-table-column label="新排课" min-width="220">
          <template #default="scope">{{ scope.row.newScheduleId ? (scope.row.newScheduleSummary || formatScheduleRef(scope.row.newScheduleId)) : '-' }}</template>
        </el-table-column>
        <el-table-column prop="operationType" label="类型" min-width="120">
          <template #default="scope">{{ formatOperationType(scope.row.operationType) }}</template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="申请时间" min-width="180" />
      </el-table>

      <el-empty v-if="!loadingList && records.length === 0" description="还没有调课申请，可从我的排课中发起第一条申请" />

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

    <el-dialog v-model="dialogVisible" title="发起调课申请" width="640px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="原排课">
          <el-select v-model="form.scheduleId" clearable filterable placeholder="请选择原排课" style="width: 100%" @change="handleSchedulePick">
            <el-option v-for="item in scheduleOptions" :key="item.id" :label="formatScheduleOption(item)" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="新上课日期">
          <el-date-picker v-model="form.newClassDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="新开始时间">
          <el-time-picker v-model="form.newStartTime" value-format="HH:mm:ss" style="width: 100%" />
        </el-form-item>
        <el-form-item label="新结束时间">
          <el-time-picker v-model="form.newEndTime" value-format="HH:mm:ss" style="width: 100%" />
        </el-form-item>
        <el-form-item label="教室">
          <el-input v-model="form.classroom" />
        </el-form-item>
        <el-form-item label="调课原因">
          <el-input v-model="form.reason" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleReschedule">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'

import { fetchRescheduleList, type RescheduleQuery, type RescheduleRecordItem } from '@/api/reschedule'
import type { ScheduleListItem } from '@/api/schedule'
import { rescheduleSchedule } from '@/api/schedule'
import { useFormOptions } from '@/composables/useFormOptions'
import { useAuthStore } from '@/stores/auth'
import { normalizeListPayload } from '@/utils/api'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'
import { cleanQueryParams } from '@/utils/query'

const route = useRoute()
const authStore = useAuthStore()
const loadingList = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)
const records = ref<RescheduleRecordItem[]>([])
const total = ref(0)
const { studentOptions, scheduleOptions, loadStudentOptions, loadScheduleOptions } = useFormOptions()

const query = reactive<RescheduleQuery>({
  studentId: '',
  operationType: '',
  pageNum: 1,
  pageSize: 10,
})

const form = reactive({
  scheduleId: null as number | null,
  newClassDate: '',
  newStartTime: '',
  newEndTime: '',
  classroom: '',
  reason: '',
})

onMounted(async () => {
  await loadStudentOptions()
  await loadScheduleOptions({ teacherId: authStore.currentUser?.id || undefined })
  if (typeof route.query.scheduleId === 'string' && Number(route.query.scheduleId) > 0) {
    form.scheduleId = Number(route.query.scheduleId)
    dialogVisible.value = true
  }
  loadList()
})

async function loadList(extra?: Partial<RescheduleQuery>) {
  loadingList.value = true
  try {
    const params = { ...query, ...extra }
    const response = await fetchRescheduleList(cleanQueryParams(params))
    const normalized = normalizeListPayload<RescheduleRecordItem>(response.data?.data)
    records.value = normalized.list
    total.value = normalized.total
    Object.assign(query, params)
  } finally {
    loadingList.value = false
  }
}

function handleSearch() {
  loadList({ pageNum: 1 })
}

function handleReset() {
  query.studentId = ''
  query.operationType = ''
  query.pageNum = 1
  query.pageSize = 10
  loadList({ pageNum: 1 })
}

function handlePageChange(pageNum: number) {
  loadList({ pageNum })
}

function handleSizeChange(pageSize: number) {
  loadList({ pageSize, pageNum: 1 })
}

function formatStudentOption(item: { id: number; studentName?: string; grade?: string; subject?: string }) {
  return [item.studentName || `未命名学员（#${item.id}）`, item.grade, item.subject].filter(Boolean).join(' · ')
}

function formatOperationType(value: string) {
  switch (value) {
    case 'leave':
      return '请假'
    case 'reschedule':
      return '调课'
    case 'makeup':
      return '补课'
    case 'cancel':
      return '取消'
    default:
      return value || '-'
  }
}

function formatScheduleOption(item: ScheduleListItem) {
  const start = String(item.startTime || '').replace('T', ' ').slice(11, 16)
  const student = item.studentName || (item.studentId ? `未命名学员（#${item.studentId}）` : '未关联学员')
  const course = item.courseName || ''
  return [item.classDate || '未排日期', start, student, course].filter(Boolean).join(' · ')
}

function formatScheduleRef(scheduleId?: number) {
  const matched = scheduleOptions.value.find((item) => item.id === scheduleId)
  if (!matched) {
    return scheduleId ? `排课记录（#${scheduleId}）` : '-'
  }
  return formatScheduleOption(matched)
}

function handleSchedulePick(scheduleId?: number) {
  const matched = scheduleOptions.value.find((item) => item.id === scheduleId)
  if (!matched) return
  form.classroom = matched.classroom || form.classroom
}

async function handleReschedule() {
  if (!form.scheduleId || !form.newClassDate || !form.newStartTime || !form.newEndTime) {
    message.warning('请完整填写原排课和新的时间安排')
    return
  }
  submitting.value = true
  try {
    await rescheduleSchedule(form.scheduleId, {
      newClassDate: form.newClassDate,
      newStartTime: form.newStartTime,
      newEndTime: form.newEndTime,
      classroom: form.classroom,
      reason: form.reason,
    })
    message.success('调课申请已提交')
    dialogVisible.value = false
    form.scheduleId = null
    form.newClassDate = ''
    form.newStartTime = ''
    form.newEndTime = ''
    form.classroom = ''
    form.reason = ''
    loadList({ pageNum: 1 })
  } catch (error) {
    message.error(extractErrorMessage(error, '调课申请提交失败，请检查排课状态与时间冲突'))
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
