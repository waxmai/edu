<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">我的学生</div>
        <h2>我的学生</h2>
        <p>查看与本人授课相关的学员档案，并快速进入排课与课堂记录。</p>
      </div>
    </div>

    <el-card shadow="never" class="toolbar-card search-form-card">
      <StudentSearchForm :model="query" @search="handleSearch" @reset="handleReset" />
    </el-card>

    <el-alert
      title="当前页面仅展示与本人授课相关的学员，可快速进入个人排课与课堂记录。"
      type="success"
      :closable="false"
    />

    <el-card shadow="never" class="section-card data-table-card">
      <el-table :data="data" v-loading="loading" border>
        <el-table-column prop="studentName" label="学员姓名" min-width="120" />
        <el-table-column prop="grade" label="年级" min-width="100" />
        <el-table-column prop="subject" label="科目" min-width="100" />
        <el-table-column prop="teachingType" label="授课形式" min-width="120">
          <template #default="scope">{{ formatTeachingType(scope.row.teachingType) }}</template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="scope">
            <StudentStatusTag :status="scope.row.status" />
          </template>
        </el-table-column>
        <el-table-column label="工作台操作" fixed="right" min-width="220">
          <template #default="scope">
            <el-space wrap>
              <el-button link type="primary" @click="goDetail(scope.row.id)">详情</el-button>
              <el-button link @click="goCreateSchedule(scope.row.id)">查看排课</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && data.length === 0" description="当前还没有分配到与您授课相关的学员" />

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
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import StudentSearchForm from '@/modules/student/components/StudentSearchForm.vue'
import StudentStatusTag from '@/modules/student/components/StudentStatusTag.vue'
import { useStudentList } from '@/modules/student/composables/useStudentList'

const router = useRouter()
const { loading, data, total, query, load, reset } = useStudentList()

onMounted(() => {
  load()
})

function handleSearch() {
  load({ pageNum: 1 })
}

function handleReset() {
  reset()
  load({ pageNum: 1 })
}

function handlePageChange(pageNum: number) {
  load({ pageNum })
}

function handleSizeChange(pageSize: number) {
  load({ pageSize, pageNum: 1 })
}

function formatTeachingType(value: string) {
  if (value === 'one_to_one') return '一对一'
  if (value === 'one_to_many') return '一对多'
  if (value === 'small_class') return '小班'
  return value || '-'
}

function goDetail(id: number) {
  router.push(`/my-students/${id}`)
}

function goCreateSchedule(id: number) {
  router.push(`/schedules?studentId=${id}&from=my-student`)
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
