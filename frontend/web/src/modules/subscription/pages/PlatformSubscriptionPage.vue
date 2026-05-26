<template>
  <div class="page-card pipeline-page" v-loading="loading">
    <div class="page-header page-header-hero platform-hero">
      <div>
        <div class="hero-kicker">平台续费运营</div>
        <h2>机构续费跟进台</h2>
        <p>集中查看即将到期、已过期、试用转化和配额逼近上限的机构，并直接维护跟进动作。</p>
      </div>
      <el-button @click="loadData">刷新</el-button>
    </div>

    <el-row :gutter="16">
      <el-col :span="6" v-for="item in summaryCards" :key="item.label">
        <el-card shadow="never" class="summary-card summary-card-clickable" @click="applySummaryFilter(item.key)">
          <div class="summary-topline">{{ item.label }}</div>
          <div class="summary-value">{{ item.value }}</div>
          <div class="summary-desc">{{ item.desc }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="content-card">
      <template #header>
        <div class="card-title">负责人跟进负载</div>
      </template>
      <div class="owner-grid">
        <div v-for="item in ownerSummaryCards" :key="item.label" class="owner-card" @click="filters.followUpOwner = item.rawOwner">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
          <small>{{ item.desc }}</small>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="content-card">
      <template #header>
        <div class="card-title">续费跟进列表</div>
      </template>

      <div class="filter-bar">
        <el-select v-model="filters.followUpStatus" clearable placeholder="跟进状态" style="width: 180px">
          <el-option label="正常维护" value="正常维护" />
          <el-option label="即将到期跟进" value="即将到期跟进" />
          <el-option label="待催续费" value="待催续费" />
          <el-option label="试用转化跟进" value="试用转化跟进" />
          <el-option label="已过期待恢复" value="已过期待恢复" />
          <el-option label="停用待处理" value="停用待处理" />
        </el-select>
        <el-input v-model="filters.followUpOwner" placeholder="负责人" clearable style="width: 180px" />
        <el-select v-model="filters.healthLevel" clearable placeholder="健康度" style="width: 140px">
          <el-option label="高风险" value="high" />
          <el-option label="关注" value="medium" />
          <el-option label="健康" value="healthy" />
        </el-select>
        <el-select v-model="filters.sortBy" placeholder="排序方式" style="width: 180px">
          <el-option label="剩余天数升序" value="remainingDaysAsc" />
          <el-option label="最近联系时间降序" value="lastContactDesc" />
          <el-option label="高风险优先" value="riskFirst" />
        </el-select>
      </div>

      <el-table :data="filteredItems" border>
        <el-table-column prop="organizationName" label="机构名称" min-width="180" />
        <el-table-column label="套餐" min-width="140">
          <template #default="scope">{{ scope.row.planName }}</template>
        </el-table-column>
        <el-table-column label="订阅状态" width="120">
          <template #default="scope">
            <el-tag :type="subscriptionTagType(scope.row.subscriptionStatus)">{{ formatSubscriptionStatus(scope.row.subscriptionStatus) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="到期时间" min-width="180">
          <template #default="scope">{{ scope.row.endsAt || '未设置' }}</template>
        </el-table-column>
        <el-table-column label="剩余天数" width="110">
          <template #default="scope">{{ scope.row.remainingDays }}</template>
        </el-table-column>
        <el-table-column label="账号配额" min-width="120">
          <template #default="scope">{{ scope.row.usedUsers }}/{{ scope.row.maxUsers }}</template>
        </el-table-column>
        <el-table-column label="校区配额" min-width="120">
          <template #default="scope">{{ scope.row.usedCampuses }}/{{ scope.row.maxCampuses }}</template>
        </el-table-column>
        <el-table-column label="健康度" width="120">
          <template #default="scope">
            <el-tag :type="healthTagType(scope.row.healthLevel)">{{ formatHealthLevel(scope.row.healthLevel) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="followUpStatus" label="跟进状态" min-width="140" />
        <el-table-column prop="followUpOwner" label="负责人" min-width="120" />
        <el-table-column prop="lastContactAt" label="最近联系" min-width="160" />
        <el-table-column prop="followUpNote" label="跟进备注" min-width="220" show-overflow-tooltip />
        <el-table-column prop="followUpRemark" label="建议动作" min-width="220" show-overflow-tooltip />
        <el-table-column label="操作" min-width="120" fixed="right">
          <template #default="scope">
            <el-button v-if="canUpdateSubscription" link type="primary" @click="openFollowUpDialog(scope.row)">维护跟进</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="维护续费跟进" width="520px">
      <el-form :model="form" label-position="top">
        <el-form-item label="跟进状态">
          <el-select v-model="form.followUpStatus" style="width: 100%">
            <el-option label="正常维护" value="正常维护" />
            <el-option label="即将到期跟进" value="即将到期跟进" />
            <el-option label="待催续费" value="待催续费" />
            <el-option label="试用转化跟进" value="试用转化跟进" />
            <el-option label="已过期待恢复" value="已过期待恢复" />
            <el-option label="停用待处理" value="停用待处理" />
          </el-select>
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.followUpOwner" placeholder="例如：销售A / 客服B" />
        </el-form-item>
        <el-form-item label="最近联系时间">
          <el-input v-model="form.lastContactAt" placeholder="2026-05-13 17:30:00" />
        </el-form-item>
        <el-form-item label="跟进备注">
          <el-input v-model="form.followUpNote" type="textarea" :rows="4" placeholder="记录最近沟通结果、承诺时间、风险点等" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitFollowUp">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { fetchSubscriptionPipeline, updateSubscriptionFollowUp, type SubscriptionPipelineItem } from '@/api/subscription'
import { usePermission } from '@/composables/usePermission'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const { can } = usePermission()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const currentOrgId = ref<number | null>(null)
const items = ref<SubscriptionPipelineItem[]>([])
const filters = reactive({
  followUpStatus: '',
  followUpOwner: '',
  healthLevel: '',
  sortBy: 'riskFirst',
})
const form = reactive({
  followUpStatus: '正常维护',
  followUpOwner: '',
  followUpNote: '',
  lastContactAt: '',
})
const canUpdateSubscription = computed(() => can({ permission: 'platform:subscription:update', menuPermission: 'menu:platform:view' }))

const summaryCards = computed(() => {
  const list = items.value
  return [
    { key: 'past_due', label: '待催续费', value: list.filter((item) => item.followUpStatus === '待催续费').length, desc: '优先触达待续费机构' },
    { key: 'expired', label: '已过期待恢复', value: list.filter((item) => item.followUpStatus === '已过期待恢复').length, desc: '尽快恢复关键业务能力' },
    { key: 'trial', label: '试用转化跟进', value: list.filter((item) => item.followUpStatus === '试用转化跟进').length, desc: '重点推进试用转正式' },
    { key: 'high', label: '高风险机构', value: list.filter((item) => item.healthLevel === 'high').length, desc: '需要优先处理的机构' },
  ]
})

const ownerSummaryCards = computed(() => {
  const ownerMap = new Map<string, number>()
  for (const item of items.value) {
    const owner = (item.followUpOwner || '未分配').trim() || '未分配'
    ownerMap.set(owner, (ownerMap.get(owner) || 0) + 1)
  }
  return [...ownerMap.entries()]
    .sort((a, b) => b[1] - a[1])
    .slice(0, 6)
    .map(([owner, count]) => ({ label: owner, rawOwner: owner === '未分配' ? '' : owner, value: count, desc: owner === '未分配' ? '建议尽快分配负责人' : '点击只看该负责人名下机构' }))
})

const filteredItems = computed(() => {
  let list = [...items.value]
  if (filters.followUpStatus) {
    list = list.filter((item) => item.followUpStatus === filters.followUpStatus)
  }
  if (filters.followUpOwner.trim()) {
    list = list.filter((item) => (item.followUpOwner || '').includes(filters.followUpOwner.trim()))
  }
  if (filters.healthLevel) {
    list = list.filter((item) => item.healthLevel === filters.healthLevel)
  }
  if (filters.sortBy === 'remainingDaysAsc') {
    list.sort((a, b) => a.remainingDays - b.remainingDays)
  } else if (filters.sortBy === 'lastContactDesc') {
    list.sort((a, b) => String(b.lastContactAt || '').localeCompare(String(a.lastContactAt || '')))
  } else {
    const weight = (item: SubscriptionPipelineItem) => (item.healthLevel === 'high' ? 0 : item.healthLevel === 'medium' ? 1 : 2)
    list.sort((a, b) => {
      const byRisk = weight(a) - weight(b)
      if (byRisk !== 0) return byRisk
      return a.remainingDays - b.remainingDays
    })
  }
  return list
})

onMounted(() => {
  filters.healthLevel = 'high'
  filters.sortBy = 'riskFirst'
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const resp = await fetchSubscriptionPipeline()
    items.value = resp.data?.data || []
  } catch (error) {
    message.error(extractErrorMessage(error, '加载续费跟进数据失败，请稍后重试'))
  } finally {
    loading.value = false
  }
}

function openFollowUpDialog(item: SubscriptionPipelineItem) {
  if (!canUpdateSubscription.value) {
    message.warning('没有维护订阅跟进权限')
    return
  }
  currentOrgId.value = item.organizationId
  form.followUpStatus = item.followUpStatus || '正常维护'
  form.followUpOwner = item.followUpOwner || ''
  form.followUpNote = item.followUpNote || ''
  form.lastContactAt = item.lastContactAt || ''
  dialogVisible.value = true
}

async function submitFollowUp() {
  if (!canUpdateSubscription.value) {
    message.warning('没有维护订阅跟进权限')
    return
  }
  if (!currentOrgId.value || !form.followUpStatus.trim()) {
    message.warning('请先填写跟进状态')
    return
  }
  submitting.value = true
  try {
    await updateSubscriptionFollowUp(currentOrgId.value, {
      followUpStatus: form.followUpStatus.trim(),
      followUpOwner: form.followUpOwner.trim(),
      followUpNote: form.followUpNote.trim(),
      lastContactAt: form.lastContactAt.trim(),
    })
    message.success('续费跟进已更新')
    dialogVisible.value = false
    await loadData()
  } catch (error) {
    message.error(extractErrorMessage(error, '保存续费跟进失败，请稍后重试'))
  } finally {
    submitting.value = false
  }
}

function applySummaryFilter(key: string) {
  if (key === 'past_due') {
    filters.followUpStatus = '待催续费'
    return
  }
  if (key === 'expired') {
    filters.followUpStatus = '已过期待恢复'
    return
  }
  if (key === 'trial') {
    filters.followUpStatus = '试用转化跟进'
    return
  }
  if (key === 'high') {
    filters.healthLevel = 'high'
  }
}

function subscriptionTagType(status?: string) {
  if (status === 'active' || status === 'trial') return 'success'
  if (status === 'past_due') return 'warning'
  return 'danger'
}

function formatSubscriptionStatus(status?: string) {
  if (status === 'trial') return '试用中'
  if (status === 'active') return '正常'
  if (status === 'past_due') return '待续费'
  if (status === 'suspended') return '已停用'
  if (status === 'expired') return '已过期'
  return status || '-'
}

function healthTagType(level?: string) {
  if (level === 'healthy') return 'success'
  if (level === 'medium') return 'warning'
  return 'danger'
}

function formatHealthLevel(level?: string) {
  if (level === 'healthy') return '健康'
  if (level === 'medium') return '关注'
  return '高风险'
}
</script>

<style scoped>
.pipeline-page {
  display: grid;
  gap: 20px;
}

.page-header-hero {
  padding: 28px 32px;
  border-radius: 24px;
}

.platform-hero {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.12), rgba(236, 72, 153, 0.08));
  border: 1px solid rgba(139, 92, 246, 0.12);
}

.hero-kicker {
  display: inline-flex;
  margin-bottom: 12px;
  padding: 7px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.06em;
  color: #7c3aed;
  background: rgba(255, 255, 255, 0.72);
}

.summary-card {
  min-height: 160px;
}

.summary-card-clickable {
  cursor: pointer;
}

.summary-topline {
  color: #667085;
  margin-bottom: 14px;
  font-size: 13px;
}

.summary-value {
  font-size: 30px;
  font-weight: 700;
  margin-bottom: 12px;
  color: #111827;
}

.summary-desc {
  color: #606266;
  line-height: 1.6;
}

.card-title {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.owner-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

.owner-card {
  padding: 16px;
  border-radius: 16px;
  border: 1px solid rgba(125, 211, 252, 0.2);
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  display: flex;
  flex-direction: column;
  gap: 6px;
  cursor: pointer;
}

.owner-card strong {
  font-size: 24px;
  color: #111827;
}

.owner-card small {
  color: #667085;
}
</style>
