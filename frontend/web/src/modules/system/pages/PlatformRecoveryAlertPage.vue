<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">恢复链路治理</div>
        <h2>恢复失败告警</h2>
        <p>查看密码找回投递失败记录，快速定位通道、目标和 challenge 状态。</p>
      </div>
      <div class="actions">
        <el-button @click="loadAlerts">刷新</el-button>
      </div>
    </div>

    <el-alert title="当前页面展示 recovery-alerts.ndjson 的受控查询结果，适合平台管理员排查验证码发送失败。" type="warning" :closable="false" />

    <div class="summary-grid">
      <el-card shadow="never" class="section-card summary-card">
        <div class="summary-label">当前筛选失败总数</div>
        <div class="summary-value">{{ summary.total }}</div>
      </el-card>
      <el-card shadow="never" class="section-card summary-card">
        <div class="summary-label">近 24 小时</div>
        <div class="summary-value">{{ summary.trend.last24Hours }}</div>
      </el-card>
      <el-card shadow="never" class="section-card summary-card">
        <div class="summary-label">近 7 天</div>
        <div class="summary-value">{{ summary.trend.last7Days }}</div>
      </el-card>
      <el-card shadow="never" class="section-card summary-card">
        <div class="summary-label">按通道</div>
        <div class="summary-tags">
          <el-tag v-for="item in channelSummaryEntries" :key="item.key" effect="light">{{ item.key }} {{ item.value }}</el-tag>
          <span v-if="channelSummaryEntries.length === 0" class="summary-empty">暂无</span>
        </div>
      </el-card>
      <el-card shadow="never" class="section-card summary-card">
        <div class="summary-label">按错误类型</div>
        <div class="summary-tags">
          <el-tag v-for="item in errorSummaryEntries" :key="item.key" type="info" effect="light">{{ item.key }} {{ item.value }}</el-tag>
          <span v-if="errorSummaryEntries.length === 0" class="summary-empty">暂无</span>
        </div>
      </el-card>
      <el-card shadow="never" class="section-card summary-card summary-card--wide">
        <div class="summary-label">近 7 天按天趋势</div>
        <div v-if="dailyTrendItems.length > 0" class="trend-chart">
          <div v-for="item in dailyTrendItems" :key="item.date" class="trend-bar-item">
            <div class="trend-bar-meta">
              <span>{{ item.shortDate }}</span>
              <strong>{{ item.count }}</strong>
            </div>
            <div class="trend-bar-track">
              <div class="trend-bar-fill" :style="{ width: item.width }"></div>
            </div>
          </div>
        </div>
        <span v-else class="summary-empty">暂无</span>
      </el-card>
      <el-card shadow="never" class="section-card summary-card summary-card--wide">
        <div class="summary-label">Top 错误文案</div>
        <div class="summary-list">
          <div v-for="item in summary.topErrors" :key="item.label" class="summary-list-item">
            <span>{{ item.label }}</span>
            <strong>{{ item.count }}</strong>
          </div>
          <span v-if="summary.topErrors.length === 0" class="summary-empty">暂无</span>
        </div>
      </el-card>
      <el-card shadow="never" class="section-card summary-card">
        <div class="summary-label">Top 用户名</div>
        <div class="summary-list">
          <div v-for="item in summary.topUsernames" :key="item.label" class="summary-list-item">
            <span>{{ item.label }}</span>
            <strong>{{ item.count }}</strong>
          </div>
          <span v-if="summary.topUsernames.length === 0" class="summary-empty">暂无</span>
        </div>
      </el-card>
      <el-card shadow="never" class="section-card summary-card">
        <div class="summary-label">Top 通道</div>
        <div class="summary-list">
          <div v-for="item in summary.topChannels" :key="item.label" class="summary-list-item">
            <span>{{ item.label }}</span>
            <strong>{{ item.count }}</strong>
          </div>
          <span v-if="summary.topChannels.length === 0" class="summary-empty">暂无</span>
        </div>
      </el-card>
    </div>

    <el-card shadow="never" class="toolbar-card search-form-card">
      <el-form :inline="true" :model="query" @submit.prevent>
        <el-form-item label="用户名">
          <el-input v-model="query.username" placeholder="用户名" clearable />
        </el-form-item>
        <el-form-item label="通道">
          <el-select v-model="query.channel" placeholder="全部通道" clearable style="width: 160px">
            <el-option label="邮箱" value="email" />
            <el-option label="短信" value="sms" />
          </el-select>
        </el-form-item>
        <el-form-item label="Challenge ID">
          <el-input v-model="query.challengeId" placeholder="challenge id" clearable style="width: 240px" />
        </el-form-item>
        <el-form-item label="目标掩码">
          <el-input v-model="query.targetMasked" placeholder="如 re***@example.local" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="section-card data-table-card" v-loading="loading">
      <el-table :data="items" border>
        <el-table-column prop="recordedAt" label="记录时间" min-width="180" />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="channel" label="通道" width="100" />
        <el-table-column prop="targetMasked" label="目标掩码" min-width="180" />
        <el-table-column prop="challengeStatus" label="Challenge 状态" width="140" />
        <el-table-column prop="expiresAt" label="过期时间" min-width="180" />
        <el-table-column prop="error" label="失败原因" min-width="260" show-overflow-tooltip />
      </el-table>

      <el-empty v-if="!loading && items.length === 0" description="暂无恢复失败记录" />

      <div class="pagination-wrap">
        <el-pagination
          background
          layout="total, prev, pager, next, sizes"
          :current-page="currentPage"
          :page-size="query.limit"
          :page-sizes="[20, 50, 100, 200]"
          :total="total"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { fetchRecoveryAlerts, fetchRecoveryAlertSummary, type RecoveryAlertItem } from '@/api/recovery-alert'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const loading = ref(false)
const items = ref<RecoveryAlertItem[]>([])
const total = ref(0)
const summary = reactive({
  total: 0,
  byChannel: {} as Record<string, number>,
  byErrorCategory: {} as Record<string, number>,
  trend: {
    last24Hours: 0,
    last7Days: 0,
  },
  topErrors: [] as Array<{ label: string; count: number }>,
  topUsernames: [] as Array<{ label: string; count: number }>,
  topChannels: [] as Array<{ label: string; count: number }>,
  daily: [] as Array<{ date: string; count: number }>,
})

const query = reactive({
  username: '',
  channel: '',
  challengeId: '',
  targetMasked: '',
  offset: 0,
  limit: 20,
})

const currentPage = computed(() => Math.floor(query.offset / query.limit) + 1)
const channelSummaryEntries = computed(() => Object.entries(summary.byChannel).map(([key, value]) => ({ key, value })))
const errorSummaryEntries = computed(() => Object.entries(summary.byErrorCategory).map(([key, value]) => ({ key, value })))
const dailyTrendItems = computed(() => {
  const max = Math.max(...summary.daily.map((item) => item.count), 0)
  return summary.daily.map((item) => ({
    ...item,
    shortDate: item.date.slice(5),
    width: max > 0 ? `${Math.max((item.count / max) * 100, item.count > 0 ? 8 : 0)}%` : '0%',
  }))
})

onMounted(() => {
  loadAlerts()
  loadSummary()
})

async function loadAlerts() {
  loading.value = true
  try {
    const { data } = await fetchRecoveryAlerts({ ...query })
    items.value = data.data?.items || []
    total.value = data.data?.total || 0
  } catch (error) {
    message.error(extractErrorMessage(error, '恢复失败告警加载失败'))
  } finally {
    loading.value = false
  }
}

async function loadSummary() {
  try {
    const { data } = await fetchRecoveryAlertSummary({
      ...query,
      offset: 0,
      limit: undefined,
    })
    summary.total = data.data?.total || 0
    summary.byChannel = data.data?.byChannel || {}
    summary.byErrorCategory = data.data?.byErrorCategory || {}
    summary.trend.last24Hours = data.data?.trend?.last24Hours || 0
    summary.trend.last7Days = data.data?.trend?.last7Days || 0
    summary.topErrors = data.data?.topErrors || []
    summary.topUsernames = data.data?.topUsernames || []
    summary.topChannels = data.data?.topChannels || []
    summary.daily = data.data?.daily || []
  } catch {
    summary.total = 0
    summary.byChannel = {}
    summary.byErrorCategory = {}
    summary.trend.last24Hours = 0
    summary.trend.last7Days = 0
    summary.topErrors = []
    summary.topUsernames = []
    summary.topChannels = []
    summary.daily = []
  }
}

function resetQuery() {
  query.username = ''
  query.channel = ''
  query.challengeId = ''
  query.targetMasked = ''
  query.offset = 0
  query.limit = 20
  loadAlerts()
  loadSummary()
}

function handleSearch() {
  query.offset = 0
  loadAlerts()
  loadSummary()
}

function handlePageChange(page: number) {
  query.offset = (page - 1) * query.limit
  loadAlerts()
  loadSummary()
}

function handleSizeChange(size: number) {
  query.limit = size
  query.offset = 0
  loadAlerts()
  loadSummary()
}
</script>

<style scoped>
.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.summary-card {
  min-height: 120px;
}

.summary-card--wide {
  grid-column: span 2;
}

.summary-label {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 10px;
}

.summary-value {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
}

.summary-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.summary-empty {
  color: #9ca3af;
  font-size: 13px;
}

.summary-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.trend-chart {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.trend-bar-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.trend-bar-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
  color: #374151;
}

.trend-bar-track {
  width: 100%;
  height: 10px;
  border-radius: 999px;
  background: #e5e7eb;
  overflow: hidden;
}

.trend-bar-fill {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #60a5fa 0%, #2563eb 100%);
}
</style>
