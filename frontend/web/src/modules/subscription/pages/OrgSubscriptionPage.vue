<template>
  <div class="page-card subscription-page" v-loading="loading">
    <div class="page-header page-header-hero">
      <div>
        <div class="hero-kicker">订阅中心</div>
        <h2>机构订阅与配额</h2>
        <p>查看当前套餐、有效期、账号与校区配额，并提前识别续费和升级窗口。</p>
      </div>
      <el-button @click="loadData">刷新</el-button>
    </div>

    <el-row :gutter="16">
      <el-col :span="6" v-for="item in summaryCards" :key="item.label">
        <el-card shadow="never" class="summary-card">
          <div class="summary-topline">{{ item.label }}</div>
          <div class="summary-value">{{ item.value }}</div>
          <div class="summary-desc">{{ item.desc }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :span="14">
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-title">套餐与有效期</div>
          </template>
          <div class="detail-grid">
            <div class="detail-item"><span>机构名称</span><strong>{{ data?.organizationName || '-' }}</strong></div>
            <div class="detail-item"><span>当前套餐</span><strong>{{ data?.planName || '-' }}</strong></div>
            <div class="detail-item"><span>交付版本</span><strong>{{ data?.editionName || '-' }}</strong></div>
            <div class="detail-item"><span>订阅状态</span><strong>{{ formatSubscriptionStatus(data?.status) }}</strong></div>
            <div class="detail-item"><span>生效时间</span><strong>{{ data?.startsAt || '-' }}</strong></div>
            <div class="detail-item"><span>到期时间</span><strong>{{ data?.endsAt || '未设置' }}</strong></div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="never" class="content-card hint-card">
          <template #header>
            <div class="card-title">续费与升级建议</div>
          </template>
          <div class="hint-block">
            <div class="hint-label">续费提醒</div>
            <p>{{ data?.renewalHint || '-' }}</p>
          </div>
          <div class="hint-block">
            <div class="hint-label">升级建议</div>
            <p>{{ data?.upgradeHint || '-' }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="content-card">
      <template #header>
        <div class="card-title">功能与资源配额</div>
      </template>
      <div class="quota-grid">
        <div class="quota-card">
          <span>账号配额</span>
          <strong>{{ data?.usedUsers || 0 }}/{{ data?.maxUsers || 0 }}</strong>
        </div>
        <div class="quota-card">
          <span>校区配额</span>
          <strong>{{ data?.usedCampuses || 0 }}/{{ data?.maxCampuses || 0 }}</strong>
        </div>
      </div>
      <div class="feature-tags">
        <el-tag v-for="flag in data?.featureFlags || []" :key="flag" size="small" effect="light">{{ formatFeatureLabel(flag) }}</el-tag>
      </div>
      <el-alert v-if="restrictedFeatures.length > 0" class="feature-alert" type="warning" :closable="false" :title="`当前套餐暂未开放：${restrictedFeatures.join('、')}`" />
    </el-card>
    <el-card shadow="never" class="content-card">
      <template #header>
        <div class="card-title">套餐权益明细</div>
      </template>
      <el-table :data="featureEntitlements" border>
        <el-table-column prop="name" label="权益" min-width="160" />
        <el-table-column label="状态" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'info'">{{ scope.row.enabled ? '已开通' : '未开通' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="280" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchCurrentSubscription, type SubscriptionOverview } from '@/api/subscription'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const loading = ref(false)
const data = ref<SubscriptionOverview | null>(null)

const featureLabelMap: Record<string, string> = {
  auth: '认证登录',
  user_management: '用户管理',
  student: '学员管理',
  course: '课程管理',
  lesson_package: '课时包管理',
  payment: '收费管理',
  schedule: '排课管理',
  lesson_record: '上课记录',
  reschedule: '调补课管理',
  platform_management: '平台管理',
  subscription_center: '订阅中心',
  audit_export: '审计导出',
  recovery_ops: '恢复治理',
}

const summaryCards = computed(() => [
  { label: '当前套餐', value: data.value?.planName || '-', desc: '当前已开通版本' },
  { label: '剩余天数', value: data.value?.remainingDays ?? 0, desc: '建议提前安排续费' },
  { label: '账号使用', value: `${data.value?.usedUsers || 0}/${data.value?.maxUsers || 0}`, desc: '当前账号配额' },
  { label: '校区使用', value: `${data.value?.usedCampuses || 0}/${data.value?.maxCampuses || 0}`, desc: '当前校区配额' },
])
const restrictedFeatures = computed(() => {
  const labels: string[] = []
  const flags = new Set(data.value?.featureFlags || [])
  if (!flags.has('subscription_center')) labels.push('订阅中心')
  if (!flags.has('audit_export')) labels.push('审计导出')
  if (!flags.has('recovery_ops')) labels.push('恢复治理')
  return labels
})

const featureEntitlements = computed(() => {
  if (data.value?.featureEntitlements?.length) {
    return data.value.featureEntitlements
  }
  const enabled = new Set(data.value?.featureFlags || [])
  return Object.entries(featureLabelMap).map(([code, name]) => ({
    code,
    name,
    enabled: enabled.has(code),
    description: '当前套餐权益',
  }))
})

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const resp = await fetchCurrentSubscription()
    data.value = resp.data?.data || null
  } catch (error) {
    message.error(extractErrorMessage(error, '加载订阅信息失败，请稍后重试'))
  } finally {
    loading.value = false
  }
}

function formatSubscriptionStatus(status?: string) {
  if (status === 'trial') return '试用中'
  if (status === 'active') return '正常'
  if (status === 'past_due') return '待续费'
  if (status === 'suspended') return '已停用'
  if (status === 'expired') return '已过期'
  return status || '-'
}

function formatFeatureLabel(flag?: string) {
  if (!flag) return '-'
  return featureLabelMap[flag] || flag
}
</script>

<style scoped>
.subscription-page {
  display: grid;
  gap: 20px;
}

.page-header-hero {
  padding: 28px 32px;
  border-radius: 24px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(16, 185, 129, 0.08));
  border: 1px solid rgba(37, 99, 235, 0.12);
}

.hero-kicker {
  display: inline-flex;
  margin-bottom: 12px;
  padding: 7px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.06em;
  color: #1d4ed8;
  background: rgba(255, 255, 255, 0.72);
}

.summary-card {
  min-height: 160px;
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

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-item span,
.hint-label {
  font-size: 13px;
  color: #667085;
}

.detail-item strong,
.quota-card strong {
  font-size: 18px;
  color: #111827;
}

.hint-card {
  height: 100%;
}

.hint-block + .hint-block {
  margin-top: 16px;
}

.hint-block p {
  margin: 8px 0 0;
  line-height: 1.7;
  color: #475467;
}

.quota-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.quota-card {
  padding: 16px;
  border-radius: 16px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  border: 1px solid rgba(37, 99, 235, 0.1);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.feature-tags {
  margin-top: 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.feature-alert {
  margin-top: 16px;
}
</style>
