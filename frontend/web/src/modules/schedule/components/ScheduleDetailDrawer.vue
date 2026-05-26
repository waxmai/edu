<template>
  <el-drawer :model-value="modelValue" title="排课详情" size="520px" @close="$emit('update:modelValue', false)">
    <el-descriptions v-if="schedule" :column="1" border>
      <el-descriptions-item label="排课ID">{{ schedule.id }}</el-descriptions-item>
      <el-descriptions-item label="学员">{{ schedule.studentName || `学员 ${schedule.studentId}` }}</el-descriptions-item>
      <el-descriptions-item label="课程">{{ schedule.courseName || `课程 ${schedule.courseId}` }}</el-descriptions-item>
      <el-descriptions-item label="教师">{{ schedule.teacherName || `教师 ${schedule.teacherId}` }}</el-descriptions-item>
      <el-descriptions-item label="课时包">{{ schedule.lessonPackageName || (schedule.lessonPackageId ? `课时包 ${schedule.lessonPackageId}` : '-') }}</el-descriptions-item>
      <el-descriptions-item label="上课日期">{{ schedule.classDate || '-' }}</el-descriptions-item>
      <el-descriptions-item label="开始时间">{{ formatDateTime(schedule.startTime) }}</el-descriptions-item>
      <el-descriptions-item label="结束时间">{{ formatDateTime(schedule.endTime) }}</el-descriptions-item>
      <el-descriptions-item label="教室">{{ schedule.classroom || '-' }}</el-descriptions-item>
      <el-descriptions-item label="状态">{{ schedule.scheduleStatus || schedule.status || '-' }}</el-descriptions-item>
      <el-descriptions-item label="备注">{{ schedule.remark || '-' }}</el-descriptions-item>
    </el-descriptions>
    <el-empty v-else description="暂未找到这条排课的详情信息" />
  </el-drawer>
</template>

<script setup lang="ts">
import type { ScheduleItem } from '@/modules/schedule/types/schedule'

defineProps<{
  modelValue: boolean
  schedule?: ScheduleItem | null
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
}>()

function formatDateTime(value?: string) {
  if (!value) {
    return '-'
  }
  return value.replace('T', ' ').slice(0, 19)
}
</script>
