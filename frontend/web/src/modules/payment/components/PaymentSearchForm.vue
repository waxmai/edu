<template>
  <el-form :inline="true" :model="model" class="payment-search-form">
    <el-form-item label="学员">
      <el-select v-model="model.studentId" clearable filterable placeholder="请选择学员" style="width: 200px" @change="handleStudentChange">
        <el-option v-for="item in studentOptions" :key="item.id" :label="formatStudentOption(item)" :value="item.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="课时包">
      <el-select v-model="model.lessonPackageId" clearable filterable placeholder="请选择课时包" style="width: 220px">
        <el-option v-for="item in lessonPackageOptions" :key="item.id" :label="lessonPackageOptionLabel(item)" :value="item.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="缴费类型">
      <el-select v-model="model.paymentType" placeholder="请选择类型" clearable style="width: 140px">
        <el-option
          v-for="option in paymentTypeOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="缴费状态">
      <el-select v-model="model.paymentStatus" placeholder="请选择状态" clearable style="width: 140px">
        <el-option
          v-for="option in paymentStatusOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="支付方式">
      <el-select v-model="model.paymentMethod" placeholder="请选择方式" clearable style="width: 140px">
        <el-option
          v-for="option in paymentMethodOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="$emit('search')">查询</el-button>
      <el-button @click="$emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { paymentMethodOptions, paymentStatusOptions, paymentTypeOptions } from '@/modules/payment/constants/payment'
import type { PaymentSearchParams } from '@/modules/payment/types/payment'
import { useFormOptions } from '@/composables/useFormOptions'

const props = defineProps<{
  model: PaymentSearchParams
}>()

defineEmits<{
  search: []
  reset: []
}>()

const { studentOptions, lessonPackageOptions, loadBaseOptions, loadLessonPackageOptions, lessonPackageOptionLabel } = useFormOptions()

function formatStudentOption(item: { id: number; studentName?: string; grade?: string; subject?: string }) {
  return [item.studentName || `学员${item.id}`, item.grade, item.subject].filter(Boolean).join(' · ')
}

onMounted(async () => {
  await loadBaseOptions()
  await loadLessonPackageOptions(typeof props.model.studentId === 'number' ? props.model.studentId : null)
})

function handleStudentChange(value?: number) {
  props.model.lessonPackageId = ''
  loadLessonPackageOptions(value || null)
}
</script>

<style scoped>
.payment-search-form {
  margin-bottom: 16px;
}
</style>
