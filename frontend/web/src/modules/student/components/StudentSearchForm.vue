<template>
  <el-form :inline="true" :model="model" class="student-search-form">
    <el-form-item label="学员姓名">
      <el-input v-model="model.studentName" placeholder="请输入学员姓名" clearable />
    </el-form-item>
    <el-form-item label="科目">
      <el-input v-model="model.subject" placeholder="请输入科目" clearable />
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="model.status" placeholder="请选择状态" clearable style="width: 140px">
        <el-option
          v-for="option in studentStatusOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="家长电话">
      <el-input v-model="model.parentPhone" placeholder="请输入家长电话" clearable />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="$emit('search')">查询</el-button>
      <el-button @click="$emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { studentStatusOptions } from '@/modules/student/constants/student'
import type { StudentSearchParams } from '@/modules/student/types/student'

defineProps<{
  model: StudentSearchParams
}>()

defineEmits<{
  search: []
  reset: []
}>()
</script>

<style scoped>
.student-search-form {
  margin-bottom: 16px;
}
</style>
