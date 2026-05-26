<template>
  <el-dialog :model-value="modelValue" title="课时包信息" width="680px" @close="$emit('update:modelValue', false)">
    <el-form :model="formModel" label-width="110px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="学员">
            <el-select v-model="formModel.studentId" filterable placeholder="请选择学员" style="width: 100%" @change="handleStudentChange">
              <el-option v-for="item in studentOptions" :key="item.id" :label="formatStudentOption(item)" :value="item.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="课程">
            <el-select v-model="formModel.courseId" filterable placeholder="请选择课程" style="width: 100%">
              <el-option v-for="item in courseOptions" :key="item.id" :label="formatCourseOption(item)" :value="item.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="总课时">
            <el-input-number v-model="formModel.totalLessons" :min="0" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="总金额">
            <el-input-number v-model="formModel.totalAmount" :min="0" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="已缴金额">
            <el-input-number v-model="formModel.paidAmount" :min="0" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="提醒阈值">
            <el-input-number v-model="formModel.lowLessonThreshold" :min="0" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="生效日期">
            <el-date-picker v-model="formModel.startDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="截止日期">
            <el-date-picker v-model="formModel.endDate" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="状态">
            <el-select v-model="formModel.status" style="width: 100%">
              <el-option
                v-for="option in lessonPackageStatusOptions.slice(1)"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="备注">
            <el-input v-model="formModel.remark" type="textarea" :rows="3" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" @click="$emit('submit', formModel)">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, watch } from 'vue'
import { lessonPackageStatusOptions } from '@/modules/lesson-package/constants/lesson-package'
import type { LessonPackageFormModel } from '@/modules/lesson-package/types/lesson-package'
import { useFormOptions } from '@/composables/useFormOptions'

const props = defineProps<{
  modelValue: boolean
  initialValue: LessonPackageFormModel
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [value: LessonPackageFormModel]
}>()

const formModel = reactive<LessonPackageFormModel>({ ...props.initialValue })
const { studentOptions, courseOptions, loadBaseOptions } = useFormOptions()

function formatStudentOption(item: { id: number; studentName?: string; grade?: string; subject?: string }) {
  return [item.studentName || `学员${item.id}`, item.grade, item.subject].filter(Boolean).join(' · ')
}

function formatCourseOption(item: { id: number; courseName?: string; subject?: string }) {
  return [item.courseName || `课程${item.id}`, item.subject].filter(Boolean).join(' · ')
}

watch(
  () => props.initialValue,
  (value) => {
    Object.assign(formModel, value)
  },
  { deep: true, immediate: true },
)

onMounted(() => {
  loadBaseOptions()
})

function handleStudentChange() {
  // reserved for future dependent filtering
}
</script>
