<template>
  <div class="page-card student-detail-page" v-loading="loading">
    <div class="page-header page-header-hero">
      <div>
        <div class="hero-kicker">我的学生</div>
        <h2>学生详情</h2>
        <p>查看与本人授课相关的学生基础信息与课堂关联记录。</p>
      </div>
      <el-space>
        <el-button type="primary" @click="goCreateSchedule">查看排课</el-button>
        <el-button @click="goBack">返回列表</el-button>
      </el-space>
    </div>

    <el-alert
      title="可从详情页查看该学生的排课和课堂记录，缴费与课时经营仅管理员可处理。"
      type="success"
      :closable="false"
    />

    <el-card shadow="never" class="section-card detail-card">
      <template v-if="detail">
        <el-descriptions :column="2" border class="detail-descriptions detail-stat-card">
          <el-descriptions-item label="学员姓名">{{ detail.studentName }}</el-descriptions-item>
          <el-descriptions-item label="性别">{{ formatGender(detail.gender) }}</el-descriptions-item>
          <el-descriptions-item label="年级">{{ detail.grade || '-' }}</el-descriptions-item>
          <el-descriptions-item label="家长姓名">{{ detail.parentName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="报读科目">{{ detail.subject || '-' }}</el-descriptions-item>
          <el-descriptions-item label="授课形式">{{ formatTeachingType(detail.teachingType) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <StudentStatusTag :status="detail.status" />
          </el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template v-else>
        <el-empty description="暂未找到这位学员的详情信息" />
      </template>
    </el-card>

    <el-row :gutter="16">
      <el-col :span="24">
        <el-card shadow="never" class="link-card detail-link-card" @click="goLessonRecords">
          <h3>课堂记录</h3>
          <p>查看该学生的课堂记录与排课关联。</p>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import StudentStatusTag from '@/modules/student/components/StudentStatusTag.vue'
import { useStudentDetail } from '@/modules/student/composables/useStudentDetail'

const route = useRoute()
const router = useRouter()
const { loading, detail, load } = useStudentDetail()

onMounted(() => {
  if (route.params.id) {
    load(String(route.params.id))
  }
})

function goBack() {
  router.push('/my-students')
}

function currentId() {
  return String(route.params.id || '')
}

function formatGender(value?: string) {
  if (value === 'male') return '男'
  if (value === 'female') return '女'
  return value || '-'
}

function formatTeachingType(value?: string) {
  if (value === 'one_to_one') return '一对一'
  if (value === 'one_to_many') return '一对多'
  if (value === 'small_class') return '小班'
  return value || '-'
}

function goCreateSchedule() {
  router.push(`/my-schedules?studentId=${currentId()}&from=my-student-detail`)
}

function goLessonRecords() {
  router.push(`/my-lesson-records?studentId=${currentId()}`)
}
</script>

<style scoped>
.student-detail-page {
  gap: 20px;
}

.detail-card :deep(.el-card__body) {
  padding-top: 18px;
}

.detail-stat-card {
  border-radius: 16px;
  overflow: hidden;
}

.detail-descriptions :deep(.el-descriptions__label) {
  font-weight: 600;
  color: #475467;
}

.detail-link-card {
  min-height: 144px;
}

h3 {
  margin-top: 0;
  margin-bottom: 10px;
  font-size: 18px;
  color: #111827;
}

.detail-link-card p {
  margin: 0;
  color: #667085;
  line-height: 1.7;
}
</style>
