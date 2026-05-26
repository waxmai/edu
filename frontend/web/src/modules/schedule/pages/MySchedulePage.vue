<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">教师工作台</div>
        <h2>我的排课</h2>
        <p>查看与本人授课相关的排课安排、调课与请假情况。</p>
      </div>
    </div>

    <el-card shadow="never" class="toolbar-card search-form-card">
      <el-form :inline="true" :model="query" @submit.prevent>
        <el-form-item label="学员">
          <el-select v-model="query.studentId" clearable filterable placeholder="请选择学员" style="width: 200px">
            <el-option v-for="item in studentOptions" :key="item.id" :label="`${item.studentName}（#${item.id}）`" :value="item.id" />
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
        <el-table-column label="操作" fixed="right" min-width="240">
          <template #default="scope">
            <el-space wrap>
              <el-button link type="primary" @click="openDetail(scope.row)">详情</el-button>
              <el-button link @click="openReschedule(scope.row)">调课</el-button>
              <el-button link type="warning" @click="openLeaveDialog(scope.row)">请假</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && data.length === 0" description="当前还没有分配到与您授课相关的排课安排" />

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

    <ScheduleDetailDrawer v-model="detailVisible" :schedule="selectedSchedule" />

    <el-dialog v-model="reasonDialogVisible" title="学员请假" width="520px">
      <el-form label-position="top">
        <el-form-item label="请假原因">
          <el-input v-model="reasonForm.reason" type="textarea" :rows="4" placeholder="请输入原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reasonDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="reasonSubmitting" @click="submitLeaveAction">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { leaveSchedule } from '@/api/schedule'
import { useFormOptions } from '@/composables/useFormOptions'
import ScheduleDetailDrawer from '@/modules/schedule/components/ScheduleDetailDrawer.vue'
import ScheduleStatusTag from '@/modules/schedule/components/ScheduleStatusTag.vue'
import { useScheduleCalendar } from '@/modules/schedule/composables/useScheduleCalendar'
import type { ScheduleListItem } from '@/api/schedule'
import { useAuthStore } from '@/stores/auth'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { loading, data, total, query, load, reset } = useScheduleCalendar()
const { studentOptions, loadStudentOptions } = useFormOptions()
const detailVisible = ref(false)
const selectedSchedule = ref<ScheduleListItem | null>(null)
const reasonDialogVisible = ref(false)
const reasonSubmitting = ref(false)
const actionScheduleId = ref<number | null>(null)
const reasonForm = reactive({ reason: '' })

onMounted(async () => {
  await loadStudentOptions()
  if (authStore.currentUser?.id) {
    query.teacherId = String(authStore.currentUser.id)
  }
  if (typeof route.query.studentId === 'string' && /^\d+$/.test(route.query.studentId.trim()) && Number(route.query.studentId) > 0) {
    query.studentId = route.query.studentId.trim()
  }
  load({ ...query })
})

function handleSearch() {
  load({ pageNum: 1 })
}

function handleReset() {
  reset()
  if (authStore.currentUser?.id) {
    query.teacherId = String(authStore.currentUser.id)
  }
  load({ pageNum: 1 })
}

function handlePageChange(pageNum: number) {
  load({ pageNum })
}

function handleSizeChange(pageSize: number) {
  load({ pageSize, pageNum: 1 })
}

function openDetail(row: ScheduleListItem) {
  selectedSchedule.value = row
  detailVisible.value = true
}

function openReschedule(row: ScheduleListItem) {
  router.push(`/my-reschedules?scheduleId=${row.id}`)
}

function openLeaveDialog(row: ScheduleListItem) {
  actionScheduleId.value = row.id
  reasonForm.reason = ''
  reasonDialogVisible.value = true
}

async function submitLeaveAction() {
  if (!actionScheduleId.value) return
  reasonSubmitting.value = true
  try {
    await leaveSchedule(actionScheduleId.value, { reason: reasonForm.reason })
    message.success('请假已登记')
    reasonDialogVisible.value = false
    load({ pageNum: 1 })
  } catch (error) {
    message.error(extractErrorMessage(error, '请假提交失败，请稍后重试'))
  } finally {
    reasonSubmitting.value = false
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
