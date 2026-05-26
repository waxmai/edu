<template>
  <div class="page-card ops-page" v-loading="loading">
    <div class="page-header page-header-hero platform-hero">
      <div>
        <div class="hero-kicker">平台运营深化</div>
        <h2>平台运营总览</h2>
        <p>集中查看租户风险、续费分配、恢复失败与后续运营深化入口。</p>
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

    <el-card shadow="never" class="content-card">
      <template #header><div class="card-title">高风险机构</div></template>
      <el-table :data="riskItems" size="small" border>
        <el-table-column prop="organizationName" label="机构名称" min-width="180" />
        <el-table-column prop="followUpStatus" label="跟进状态" min-width="140" />
        <el-table-column prop="remainingDays" label="剩余天数" width="100" />
        <el-table-column label="账号配额" min-width="120">
          <template #default="scope">{{ scope.row.usedUsers }}/{{ scope.row.maxUsers }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card shadow="never" class="content-card">
      <template #header><div class="card-title">修复与排查入口</div></template>
      <div class="quick-grid">
        <button class="quick-entry" type="button" @click="go('/platform/recovery-alerts')">
          <span class="quick-entry-title">恢复异常排查</span>
          <span class="quick-entry-desc">从恢复失败告警页进入 challenge / audit 相关排查</span>
        </button>
        <button class="quick-entry" type="button" @click="go('/platform/audit-logs')">
          <span class="quick-entry-title">审计追溯</span>
          <span class="quick-entry-desc">按 Trace ID 和动作回放关键操作，辅助修复判断</span>
        </button>
        <button class="quick-entry" type="button" @click="go('/platform/tenants')">
          <span class="quick-entry-title">租户治理</span>
          <span class="quick-entry-desc">先核验租户状态、配额、设置与健康级别，再判断是否需要进一步修复</span>
        </button>
      </div>
      <div class="repair-tips">
        <div class="repair-tips-title">当前建议排查顺序</div>
        <ol class="repair-steps">
          <li>先看异常入口定位范围</li>
          <li>再用审计日志回放 Trace</li>
          <li>确认是配置、权限、数据还是投递链路问题</li>
          <li>最后再决定是否进入人工修复</li>
        </ol>
      </div>
    </el-card>

    <el-card shadow="never" class="content-card">
      <template #header><div class="card-title">运营深化入口</div></template>
      <div class="quick-grid">
        <button class="quick-entry" type="button" @click="go('/platform/subscriptions')">
          <span class="quick-entry-title">续费跟进台</span>
          <span class="quick-entry-desc">处理即将到期、待催续费与负责人分配</span>
        </button>
        <button class="quick-entry" type="button" @click="go('/platform/recovery-alerts')">
          <span class="quick-entry-title">恢复失败治理</span>
          <span class="quick-entry-desc">排查恢复投递失败、趋势与错误聚类</span>
        </button>
        <button class="quick-entry" type="button" @click="go('/platform/tenants')">
          <span class="quick-entry-title">租户治理</span>
          <span class="quick-entry-desc">查看高风险租户、配额与租户设置状态</span>
        </button>
        <button class="quick-entry" type="button" @click="go('/platform/audit-logs')">
          <span class="quick-entry-title">审计追溯</span>
          <span class="quick-entry-desc">回放关键操作，辅助平台安全排障</span>
        </button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchDashboardData, type DashboardResponse } from '@/api/dashboard'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const router = useRouter()
const loading = ref(false)
const data = ref<DashboardResponse>({ role: '', summaryCards: [] })

const summaryCards = computed(() => {
  const metrics = new Map((data.value.metricItems || []).map((item) => [item.label, item.value]))
  return [
    { label: '高风险机构', value: data.value.platformRisks?.length || 0, desc: '当前需优先处理的高风险租户' },
    { label: '机构总数', value: metrics.get('机构总数') || '-', desc: '平台已开通机构规模' },
    { label: '总账号数', value: metrics.get('总账号数') || '-', desc: '平台账号总体规模' },
    { label: '总校区数', value: metrics.get('总校区数') || '-', desc: '平台校区总体规模' },
  ]
})
const riskItems = computed(() => data.value.platformRisks || [])

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const resp = await fetchDashboardData()
    data.value = resp.data?.data || { role: '', summaryCards: [] }
  } catch (error) {
    message.error(extractErrorMessage(error, '加载平台运营总览失败'))
  } finally {
    loading.value = false
  }
}

function go(path: string) {
  router.push(path)
}
</script>

<style scoped>
.ops-page { display: grid; gap: 20px; }
.page-header-hero { padding: 28px 32px; border-radius: 24px; }
.platform-hero { background: linear-gradient(135deg, rgba(139, 92, 246, 0.12), rgba(236, 72, 153, 0.08)); border: 1px solid rgba(139, 92, 246, 0.12); }
.hero-kicker { display: inline-flex; margin-bottom: 12px; padding: 7px 12px; border-radius: 999px; font-size: 12px; font-weight: 700; color: #7c3aed; background: rgba(255,255,255,0.72); }
.summary-card { min-height: 160px; }
.summary-topline { color: #667085; margin-bottom: 14px; font-size: 13px; }
.summary-value { font-size: 30px; font-weight: 700; margin-bottom: 12px; color: #111827; }
.summary-desc { color: #606266; line-height: 1.6; }
.card-title { font-size: 16px; font-weight: 700; color: #111827; }
.quick-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 14px; }
.quick-entry { border: 1px solid rgba(15, 23, 42, 0.08); background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%); border-radius: 18px; padding: 18px; text-align: left; cursor: pointer; }
.quick-entry-title { display: block; font-size: 15px; font-weight: 700; color: #111827; }
.quick-entry-desc { display: block; margin-top: 8px; font-size: 13px; line-height: 1.6; color: #667085; }
.repair-tips { margin-top: 16px; padding: 16px 18px; border-radius: 16px; background: #f8fafc; border: 1px solid rgba(148, 163, 184, 0.18); }
.repair-tips-title { font-size: 14px; font-weight: 700; color: #0f172a; margin-bottom: 10px; }
.repair-steps { margin: 0; padding-left: 18px; color: #475569; line-height: 1.8; font-size: 13px; }
</style>
