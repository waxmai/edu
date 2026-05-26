<template>
  <div class="page-card settings-page" v-loading="loading">
    <div class="page-header page-header-hero">
      <div>
        <div class="hero-kicker">租户设置中心</div>
        <h2>机构设置</h2>
        <p>维护机构品牌、通知、安全策略和运营备注，作为租户侧设置中心的基础版入口。</p>
      </div>
      <el-button @click="loadData">刷新</el-button>
    </div>

    <el-row :gutter="16">
      <el-col :span="8">
        <el-card shadow="never" class="summary-card">
          <div class="summary-topline">当前套餐</div>
          <div class="summary-value">{{ settings?.planCode || '-' }}</div>
          <div class="summary-desc">受 feature flags 与订阅状态共同约束</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" class="summary-card">
          <div class="summary-topline">订阅状态</div>
          <div class="summary-value">{{ formatSubscriptionStatus(settings?.subscriptionStatus) }}</div>
          <div class="summary-desc">设置变更同样受订阅状态保护</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" class="summary-card">
          <div class="summary-topline">功能开关数</div>
          <div class="summary-value">{{ settings?.featureFlags?.length || 0 }}</div>
          <div class="summary-desc">当前租户已开通功能范围</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :span="14">
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-title">基础设置</div>
          </template>
          <el-form label-position="top">
            <el-form-item label="机构名称">
              <el-input :model-value="settings?.organizationName || '-'" disabled />
            </el-form-item>
            <el-form-item label="时区">
              <el-input v-model="form.timezone" placeholder="Asia/Shanghai" />
            </el-form-item>
            <el-form-item label="品牌名称">
              <el-input v-model="form.brandName" placeholder="如：明皇教培" />
            </el-form-item>
            <el-form-item label="通知邮箱">
              <el-input v-model="form.notificationEmail" placeholder="ops@example.com" />
            </el-form-item>
            <el-form-item label="安全策略">
              <el-select v-model="form.securityPolicy" multiple collapse-tags style="width: 100%" placeholder="选择安全策略">
                <el-option label="强制复杂密码" value="strong_password" />
                <el-option label="异常登录提醒" value="login_alert" />
                <el-option label="定期审计复核" value="audit_review" />
              </el-select>
            </el-form-item>
            <el-form-item label="运营备注">
              <el-input v-model="form.remark" type="textarea" :rows="4" placeholder="可记录租户运营习惯、交付注意事项、协作说明等" />
            </el-form-item>
            <el-button type="primary" :loading="saving" @click="submit">保存设置</el-button>
          </el-form>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-title">已开通功能与安全策略</div>
          </template>
          <div class="feature-tags">
            <el-tag v-for="flag in settings?.featureFlags || []" :key="flag" size="small" effect="light">{{ formatFeatureLabel(flag) }}</el-tag>
          </div>
          <div class="policy-tags">
            <el-tag v-for="policy in settings?.securityPolicy || []" :key="policy" type="warning" effect="light">{{ formatPolicyLabel(policy) }}</el-tag>
            <span v-if="!settings?.securityPolicy?.length" class="policy-empty">暂未配置</span>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { fetchOrganizationSettings, updateOrganizationSettings, type OrganizationSettings } from '@/api/platform'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const loading = ref(false)
const saving = ref(false)
const settings = ref<OrganizationSettings | null>(null)
const form = reactive({
  timezone: 'Asia/Shanghai',
  brandName: '',
  notificationEmail: '',
  securityPolicy: [] as string[],
  remark: '',
})

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

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const resp = await fetchOrganizationSettings()
    settings.value = resp.data?.data || null
    form.timezone = settings.value?.timezone || 'Asia/Shanghai'
    form.brandName = settings.value?.brandName || ''
    form.notificationEmail = settings.value?.notificationEmail || ''
    form.securityPolicy = settings.value?.securityPolicy || []
    form.remark = settings.value?.remark || ''
  } catch (error) {
    message.error(extractErrorMessage(error, '加载机构设置失败，请稍后重试'))
  } finally {
    loading.value = false
  }
}

async function submit() {
  saving.value = true
  try {
    await updateOrganizationSettings({
      timezone: form.timezone.trim(),
      brandName: form.brandName.trim(),
      notificationEmail: form.notificationEmail.trim(),
      securityPolicy: form.securityPolicy,
      remark: form.remark.trim(),
    })
    message.success('机构设置已保存')
    await loadData()
  } catch (error) {
    message.error(extractErrorMessage(error, '保存机构设置失败，请稍后重试'))
  } finally {
    saving.value = false
  }
}

function formatFeatureLabel(flag?: string) {
  if (!flag) return '-'
  return featureLabelMap[flag] || flag
}

function formatPolicyLabel(policy?: string) {
  if (policy === 'strong_password') return '强制复杂密码'
  if (policy === 'login_alert') return '异常登录提醒'
  if (policy === 'audit_review') return '定期审计复核'
  return policy || '-'
}

function formatSubscriptionStatus(status?: string) {
  if (status === 'trial') return '试用中'
  if (status === 'active') return '正常'
  if (status === 'past_due') return '待续费'
  if (status === 'suspended') return '已停用'
  if (status === 'expired') return '已过期'
  return status || '-'
}
</script>

<style scoped>
.settings-page { display: grid; gap: 20px; }
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
.summary-card { min-height: 160px; }
.summary-topline { color: #667085; margin-bottom: 14px; font-size: 13px; }
.summary-value { font-size: 28px; font-weight: 700; margin-bottom: 12px; color: #111827; }
.summary-desc { color: #606266; line-height: 1.6; }
.card-title { font-size: 16px; font-weight: 700; color: #111827; }
.feature-tags { display: flex; flex-wrap: wrap; gap: 8px; }
.policy-tags { margin-top: 12px; display: flex; flex-wrap: wrap; gap: 8px; }
.policy-empty { color: #667085; font-size: 13px; }
</style>
