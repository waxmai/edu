<template>
  <el-dialog :model-value="modelValue" :title="title" width="760px" @close="$emit('update:modelValue', false)">
    <el-form :model="formModel" label-width="110px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="学员" required>
            <el-select v-model="formModel.studentId" filterable placeholder="请选择学员" style="width: 100%" @change="handleStudentChange">
              <el-option v-for="item in studentOptions" :key="item.id" :label="formatStudentOption(item)" :value="item.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="课程" required>
            <el-select v-model="formModel.courseId" filterable placeholder="请选择课程" style="width: 100%">
              <el-option v-for="item in courseOptions" :key="item.id" :label="formatCourseOption(item)" :value="item.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="教师" required>
            <el-select v-model="formModel.teacherId" filterable placeholder="请选择教师" style="width: 100%">
              <el-option v-for="item in teacherOptions" :key="item.id" :label="formatTeacherOption(item)" :value="item.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="课时包">
            <el-select v-model="formModel.lessonPackageId" clearable filterable placeholder="请选择课时包" style="width: 100%">
              <el-option v-for="item in lessonPackageOptions" :key="item.id" :label="lessonPackageOptionLabel(item)" :value="item.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="上课日期" required>
            <el-date-picker v-model="formModel.classDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="教室">
            <el-input v-model="formModel.classroom" placeholder="如：教室A / 线上会议室" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="开始时间" required>
            <el-time-picker v-model="formModel.startTime" value-format="HH:mm:ss" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="结束时间" required>
            <el-time-picker v-model="formModel.endTime" value-format="HH:mm:ss" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="排课状态">
            <el-select v-model="formModel.scheduleStatus" style="width: 100%">
              <el-option v-for="option in scheduleStatusOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="备注">
            <el-input v-model="formModel.remark" type="textarea" :rows="3" placeholder="可填写教室说明、排课备注等" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" @click="handleSubmit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, watch } from 'vue'
import { scheduleStatusOptions } from '@/modules/schedule/constants/schedule'
import type { ScheduleCreatePayload } from '@/api/schedule'
import { useFormOptions } from '@/composables/useFormOptions'

interface ScheduleFormValue {
  studentId: number | null
  courseId: number | null
  teacherId: number | null
  lessonPackageId: number | null
  classDate: string
  startTime: string
  endTime: string
  classroom: string
  scheduleStatus: string
  remark: string
}

const props = withDefaults(defineProps<{
  modelValue: boolean
  initialValue: ScheduleFormValue
  title?: string
}>(), {
  title: '排课信息',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [value: ScheduleCreatePayload]
}>()

const formModel = reactive<ScheduleFormValue>({ ...props.initialValue })
const { studentOptions, courseOptions, lessonPackageOptions, teacherOptions, loadBaseOptions, loadLessonPackageOptions, lessonPackageOptionLabel } = useFormOptions()

function formatStudentOption(item: { id: number; studentName?: string; grade?: string; subject?: string }) {
  return [item.studentName || `学员${item.id}`, item.grade, item.subject].filter(Boolean).join(' · ')
}

function formatCourseOption(item: { id: number; courseName?: string; subject?: string }) {
  return [item.courseName || `课程${item.id}`, item.subject].filter(Boolean).join(' · ')
}

function formatTeacherOption(item: { id: number; realName?: string; username?: string }) {
  return [item.realName || item.username || `教师${item.id}`, item.username ? `@${item.username}` : ''].filter(Boolean).join(' · ')
}

watch(
  () => props.initialValue,
  (value) => {
    Object.assign(formModel, value)
    if (value.studentId) {
      loadLessonPackageOptions(value.studentId)
    }
  },
  { deep: true, immediate: true },
)

onMounted(() => {
  loadBaseOptions()
  if (formModel.studentId) {
    loadLessonPackageOptions(formModel.studentId)
  }
})

const payload = computed(() => ({
  studentId: formModel.studentId,
  courseId: formModel.courseId,
  teacherId: formModel.teacherId,
  lessonPackageId: formModel.lessonPackageId || null,
  classDate: formModel.classDate,
  startTime: formModel.classDate && formModel.startTime ? `${formModel.classDate} ${formModel.startTime}` : '',
  endTime: formModel.classDate && formModel.endTime ? `${formModel.classDate} ${formModel.endTime}` : '',
  classroom: formModel.classroom,
  scheduleStatus: formModel.scheduleStatus || 'scheduled',
  remark: formModel.remark,
}))

function handleStudentChange(value: number) {
  formModel.lessonPackageId = null
  loadLessonPackageOptions(value)
}

function handleSubmit() {
  if (!formModel.studentId || !formModel.courseId || !formModel.teacherId) {
    return
  }
  const submitPayload: ScheduleCreatePayload = {
    studentId: formModel.studentId,
    courseId: formModel.courseId,
    teacherId: formModel.teacherId,
    lessonPackageId: formModel.lessonPackageId || null,
    classDate: payload.value.classDate,
    startTime: payload.value.startTime,
    endTime: payload.value.endTime,
    classroom: payload.value.classroom,
    scheduleStatus: payload.value.scheduleStatus,
    remark: payload.value.remark,
  }
  emit('submit', submitPayload)
}
</script>
