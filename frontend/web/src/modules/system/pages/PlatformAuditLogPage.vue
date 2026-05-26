<template>
  <div class="page-card">
    <div class="page-header">
      <div>
        <div class="hero-kicker">安全与追溯</div>
        <h2>审计日志</h2>
        <p>按动作、模块、操作人、Trace ID 与时间范围快速追查关键操作。</p>
      </div>
      <div class="actions">
        <el-dropdown v-if="canExportAudit" trigger="click" @command="handleExportCommand">
          <el-button :loading="exporting">
            导出
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="page">导出当前页</el-dropdown-item>
              <el-dropdown-item command="all">导出全部匹配（最多 5000 条）</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button @click="copyTraceId" :disabled="!selectedLog?.traceId">复制 Trace ID</el-button>
        <el-button @click="loadLogs">刷新</el-button>
      </div>
    </div>

    <el-alert title="当前页面仅展示结构化审计日志，适合平台管理员做受控排查。" type="success" :closable="false" />
    <el-alert title="导出支持当前页和全部匹配结果，全部匹配单次最多导出 5000 条。" type="info" :closable="false" class="mt-12" />

    <el-card shadow="never" class="toolbar-card search-form-card">
      <el-form :inline="true" :model="query" @submit.prevent>
        <el-form-item label="动作">
          <el-select v-model="query.action" placeholder="全部动作" clearable filterable style="width: 220px">
            <el-option v-for="item in actionOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="模块">
          <el-select v-model="query.module" placeholder="全部模块" clearable filterable style="width: 180px">
            <el-option v-for="item in moduleOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作人">
          <el-input v-model="query.actorUsername" placeholder="用户名" clearable />
        </el-form-item>
        <el-form-item label="目标对象">
          <el-select
            v-model="query.targetId"
            placeholder="搜索名称 / ID"
            clearable
            filterable
            remote
            reserve-keyword
            :remote-method="searchTargets"
            :loading="targetSearching"
            style="width: 260px"
          >
            <el-option v-for="item in targetOptions" :key="`${item.module}-${item.id}`" :label="item.label" :value="item.id">
              <div class="target-option">
                <span>{{ item.label }}</span>
                <small v-if="item.extra">{{ item.extra }}</small>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="Trace ID">
          <el-input v-model="query.traceId" placeholder="追踪链路" clearable style="width: 240px" />
        </el-form-item>
        <el-form-item label="开始时间">
          <el-date-picker v-model="query.startTime" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" placeholder="开始时间" />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-date-picker v-model="query.endTime" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" placeholder="结束时间" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadLogs">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="section-card data-table-card" v-loading="loading">
      <el-table :data="logs" border @row-click="selectRow">
        <el-table-column prop="time" label="时间" min-width="180" />
        <el-table-column prop="action" label="动作" min-width="180" />
        <el-table-column prop="module" label="模块" min-width="120" />
        <el-table-column prop="actorUsername" label="操作人" min-width="140" />
        <el-table-column prop="actorRole" label="角色" min-width="140" />
        <el-table-column prop="targetId" label="目标ID" width="100" />
        <el-table-column prop="traceId" label="Trace ID" min-width="220" show-overflow-tooltip />
      </el-table>

      <el-empty v-if="!loading && logs.length === 0" description="暂无匹配的审计记录" />

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

    <el-card shadow="never" class="section-card" v-if="selectedLog">
      <template #header>
        <div class="detail-header">
          <span>审计详情</span>
          <el-tag type="info">{{ selectedLog.action }}</el-tag>
        </div>
      </template>
      <div class="detail-grid">
        <div><strong>时间：</strong>{{ selectedLog.time || '-' }}</div>
        <div><strong>模块：</strong>{{ selectedLog.module || '-' }}</div>
        <div><strong>操作人：</strong>{{ selectedLog.actorUsername || '-' }}</div>
        <div><strong>角色：</strong>{{ selectedLog.actorRole || '-' }}</div>
        <div><strong>动作：</strong>{{ selectedLog.action || '-' }}</div>
        <div><strong>目标ID：</strong>{{ selectedLog.targetId || '-' }}</div>
        <div><strong>Trace ID：</strong>{{ selectedLog.traceId || '-' }}</div>
      </div>
      <el-divider />
      <div class="detail-summary">
        <div class="detail-summary-title">关键字段</div>
        <div v-if="detailEntries.length === 0" class="detail-summary-empty">当前记录没有 detail 字段</div>
        <el-descriptions v-else :column="1" border size="small">
          <el-descriptions-item v-for="item in detailEntries" :key="item.key" :label="item.key">
            {{ item.value }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <el-divider />
      <pre class="detail-json">{{ prettyDetail }}</pre>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ArrowDown } from '@element-plus/icons-vue'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { exportAuditLogs, fetchAuditLogMeta, fetchAuditLogs, searchAuditTargets, type AuditLogItem, type AuditTargetOption } from '@/api/audit'
import { usePermission } from '@/composables/usePermission'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const { can } = usePermission()
const loading = ref(false)
const exporting = ref(false)
const logs = ref<AuditLogItem[]>([])
const total = ref(0)
const selectedLog = ref<AuditLogItem | null>(null)
const canExportAudit = computed(() => can({ permission: 'platform:audit:export', menuPermission: 'menu:platform:view' }))
const actionOptions = ref<string[]>([])
const moduleOptions = ref<string[]>([])
const targetOptions = ref<AuditTargetOption[]>([])
const targetSearching = ref(false)

const query = reactive({
  action: '',
  module: '',
  actorUsername: '',
  targetId: undefined as number | undefined,
  traceId: '',
  startTime: '',
  endTime: '',
  offset: 0,
  limit: 20,
})

const currentPage = computed(() => Math.floor(query.offset / query.limit) + 1)
const prettyDetail = computed(() => JSON.stringify(selectedLog.value?.detail || {}, null, 2))
const detailEntries = computed(() => {
  const detail = selectedLog.value?.detail || {}
  return Object.entries(detail).map(([key, value]) => ({ key, value: typeof value === 'string' ? value : JSON.stringify(value) }))
})

onMounted(() => {
  loadMeta()
  searchTargets('')
  loadLogs()
})

watch(
  () => query.module,
  () => {
    query.targetId = undefined
    targetOptions.value = []
    searchTargets('')
  },
)

async function loadMeta() {
  try {
    const { data } = await fetchAuditLogMeta()
    actionOptions.value = data.data?.actions || []
    moduleOptions.value = data.data?.modules || []
  } catch {
    actionOptions.value = []
    moduleOptions.value = []
  }
}

async function searchTargets(keyword: string) {
  targetSearching.value = true
  try {
    const { data } = await searchAuditTargets({ module: query.module, keyword, limit: 20 })
    targetOptions.value = data.data?.items || []
  } catch {
    targetOptions.value = []
  } finally {
    targetSearching.value = false
  }
}

async function loadLogs() {
  loading.value = true
  try {
    const { data } = await fetchAuditLogs({ ...query })
    const payload = data.data
    logs.value = payload?.items || []
    total.value = payload?.total || 0
    selectedLog.value = logs.value.length > 0 ? logs.value[0] : null
  } catch (error) {
    message.error(extractErrorMessage(error, '审计日志加载失败'))
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.action = ''
  query.module = ''
  query.actorUsername = ''
  query.targetId = undefined
  targetOptions.value = []
  searchTargets('')
  query.traceId = ''
  query.startTime = ''
  query.endTime = ''
  query.offset = 0
  query.limit = 20
  loadLogs()
}

function handlePageChange(page: number) {
  query.offset = (page - 1) * query.limit
  loadLogs()
}

function handleSizeChange(size: number) {
  query.limit = size
  query.offset = 0
  loadLogs()
}

function selectRow(row: AuditLogItem) {
  selectedLog.value = row
}

async function copyTraceId() {
  const traceId = selectedLog.value?.traceId
  if (!traceId) return
  try {
    await navigator.clipboard.writeText(traceId)
    message.success('Trace ID 已复制')
  } catch {
    message.error('复制 Trace ID 失败')
  }
}

async function exportLogs(mode: 'page' | 'all') {
  if (!canExportAudit.value) {
    message.warning('没有导出审计日志权限')
    return
  }
  exporting.value = true
  try {
    const pageLimit = query.limit
    const pageOffset = query.offset
    const exportLimit = mode === 'page' ? pageLimit : total.value > 0 ? Math.min(total.value, 5000) : 1000
    const blob = await exportAuditLogs({
      ...query,
      offset: mode === 'page' ? pageOffset : 0,
      limit: exportLimit,
    })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = buildExportFilename(mode)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    if (mode === 'all' && total.value > 5000) {
      message.success('审计日志已导出，已按上限导出前 5000 条')
      return
    }
    message.success(mode === 'page' ? '当前页审计日志导出成功' : '全部匹配审计日志导出成功')
  } catch (error) {
    message.error(extractErrorMessage(error, '审计日志导出失败'))
  } finally {
    exporting.value = false
  }
}

function handleExportCommand(command: string) {
  if (command === 'page' || command === 'all') {
    exportLogs(command)
  }
}

function buildExportFilename(mode: 'page' | 'all') {
  const scope = mode === 'page' ? `page-${currentPage.value}` : `all-${Math.min(total.value || 0, 5000) || 0}`
  return `audit-logs-${scope}-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')}.csv`
}
</script>

<style scoped>
.target-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.target-option small {
  color: #909399;
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 12px;
}

.detail-summary {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-summary-title {
  font-weight: 600;
  color: #1f2937;
}

.detail-summary-empty {
  color: #6b7280;
  font-size: 13px;
}

.detail-json {
  margin: 0;
  padding: 16px;
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 12px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.5;
}
</style>
