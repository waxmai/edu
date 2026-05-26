<template>
  <el-form :inline="true" :model="model" class="lesson-package-search-form">
    <el-form-item label="学员">
      <el-select v-model="model.studentId" clearable filterable placeholder="请选择学员" style="width: 200px">
        <el-option v-for="item in studentOptions" :key="item.id" :label="`${item.studentName}（#${item.id}）`" :value="item.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="课程">
      <el-select v-model="model.courseId" clearable filterable placeholder="请选择课程" style="width: 200px">
        <el-option v-for="item in courseOptions" :key="item.id" :label="`${item.courseName}（#${item.id}）`" :value="item.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="model.status" placeholder="请选择状态" clearable style="width: 140px">
        <el-option
          v-for="option in lessonPackageStatusOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="emitSearch">查询</el-button>
      <el-button @click="$emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { lessonPackageStatusOptions } from '@/modules/lesson-package/constants/lesson-package'
import type { LessonPackageSearchParams } from '@/modules/lesson-package/types/lesson-package'
import { useFormOptions } from '@/composables/useFormOptions'

const props = defineProps<{
  model: LessonPackageSearchParams
}>()

const emit = defineEmits<{
  search: []
  reset: []
}>()

const { studentOptions, courseOptions, loadBaseOptions } = useFormOptions()

onMounted(() => {
  loadBaseOptions()
})

function emitSearch() {
  emit('search')
}
</script>

<style scoped>
.lesson-package-search-form {
  margin-bottom: 16px;
}
</style>
