<template>
  <el-dialog :model-value="modelValue" title="缴费记录" width="640px" @close="$emit('update:modelValue', false)">
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
          <el-form-item label="课时包">
            <el-select v-model="formModel.lessonPackageId" filterable placeholder="请选择课时包" style="width: 100%">
              <el-option v-for="item in lessonPackageOptions" :key="item.id" :label="lessonPackageOptionLabel(item)" :value="item.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="缴费类型">
            <el-select v-model="formModel.paymentType" style="width: 100%">
              <el-option
                v-for="option in paymentTypeOptions.slice(1)"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="缴费金额">
            <el-input-number v-model="formModel.amount" :min="0" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="支付方式">
            <el-select v-model="formModel.paymentMethod" style="width: 100%">
              <el-option
                v-for="option in paymentMethodOptions.slice(1)"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="缴费状态">
            <el-select v-model="formModel.paymentStatus" style="width: 100%">
              <el-option
                v-for="option in paymentStatusOptions.slice(1)"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="缴费时间">
            <el-date-picker
              v-model="formModel.paymentTime"
              type="datetime"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%"
            />
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
import { paymentMethodOptions, paymentStatusOptions, paymentTypeOptions } from '@/modules/payment/constants/payment'
import type { PaymentFormModel } from '@/modules/payment/types/payment'
import { useFormOptions } from '@/composables/useFormOptions'

const props = defineProps<{
  modelValue: boolean
  initialValue: PaymentFormModel
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [value: PaymentFormModel]
}>()

const formModel = reactive<PaymentFormModel>({ ...props.initialValue })
const { studentOptions, lessonPackageOptions, loadBaseOptions, loadLessonPackageOptions, lessonPackageOptionLabel } = useFormOptions()

function formatStudentOption(item: { id: number; studentName?: string; grade?: string; subject?: string }) {
  return [item.studentName || `学员${item.id}`, item.grade, item.subject].filter(Boolean).join(' · ')
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

function handleStudentChange(value: number) {
  formModel.lessonPackageId = null
  loadLessonPackageOptions(value)
}
</script>
