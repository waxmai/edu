<template>
  <div class="login-page">
    <div class="login-hero">
      <div class="login-hero-badge">面向试点与商用的运营后台</div>
      <h1>让排课、课包、收费与学员管理落在一个清晰的工作台里</h1>
      <p>面向真实教培业务的日常运营场景，统一处理排课安排、课时消耗、收费管理与人员协作。</p>
      <div class="login-feature-list">
        <span>统一业务主链</span>
        <span>正式鉴权与角色控制</span>
        <span>适合受控商业化部署</span>
      </div>
    </div>

    <el-card class="login-card" shadow="hover">
      <div class="login-header">
        <h1>欢迎登录</h1>
        <p>请输入正式账号和密码进入系统。</p>
      </div>

      <el-form label-position="top" :model="form" @submit.prevent>
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" @keyup.enter="handleLogin" />
        </el-form-item>

        <el-space fill style="width: 100%">
          <el-button type="primary" :loading="submitting" @click="handleLogin">登录系统</el-button>
        </el-space>
      </el-form>

      <div class="login-footer">
        登录成功后将调用 <code>/api/v1/auth/me</code> 初始化当前用户，并在本地持久化 token。若账号为首次登录状态，系统会进入首页并立即弹出阻断式改密窗口。
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { login } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { openPasswordDialog } from '@/stores/password-dialog'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const submitting = ref(false)

const form = reactive({
  username: '',
  password: '',
})

async function handleLogin() {
  if (!form.username.trim() || !form.password.trim()) {
    message.error('请填写用户名和密码')
    return
  }

  submitting.value = true
  try {
    const response = await login({
      username: form.username.trim(),
      password: form.password,
    })
    const payload = response.data?.data
    if (!payload?.token || !payload.userInfo) {
      throw new Error('登录响应缺少会话信息')
    }

    authStore.setToken(payload.token)
    authStore.setRefreshToken(payload.refreshToken)
    authStore.setCurrentUser(payload.userInfo)
    authStore.bootstrapped = true

    message.success(payload.userInfo.mustChangePassword ? '登录成功，请先修改密码' : '登录成功')
    const redirect = typeof route.query.redirect === 'string'
      ? route.query.redirect
      : '/dashboard'
    await router.push(redirect)
    if (payload.userInfo.mustChangePassword) {
      authStore.activateForcePasswordChange()
      openPasswordDialog(true)
    }
  } catch (error: any) {
    console.error(error)
    message.error(extractErrorMessage(error, '登录失败，请检查账号状态或密码后重试'))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(360px, 1.05fr) minmax(420px, 520px);
  align-items: center;
  gap: 56px;
  padding: 48px 72px;
  background:
    radial-gradient(circle at top left, rgba(37, 99, 235, 0.16), transparent 26%),
    radial-gradient(circle at bottom right, rgba(16, 185, 129, 0.1), transparent 24%),
    linear-gradient(135deg, #f2f7ff 0%, #f8fbff 100%);
}

.login-hero {
  max-width: 620px;
}

.login-hero-badge {
  display: inline-flex;
  align-items: center;
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.08);
  color: #1d4ed8;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
  margin-bottom: 20px;
}

.login-hero h1 {
  margin: 0;
  font-size: 44px;
  line-height: 1.15;
  letter-spacing: -0.05em;
  color: #0f172a;
}

.login-hero p {
  margin: 18px 0 0;
  font-size: 16px;
  line-height: 1.8;
  color: #667085;
  max-width: 560px;
}

.login-feature-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 26px;
}

.login-feature-list span {
  padding: 10px 14px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(15, 23, 42, 0.08);
  color: #334155;
  font-size: 13px;
}

.login-card {
  width: 100%;
  max-width: 100%;
  border-radius: 24px;
  overflow: hidden;
}

.login-header {
  margin-bottom: 20px;
}

.login-header h1 {
  margin: 0 0 8px;
  font-size: 30px;
  letter-spacing: -0.03em;
}

.login-header p {
  margin: 0;
  color: #606266;
}

.login-footer {
  margin-top: 16px;
  color: #606266;
  line-height: 1.7;
  font-size: 13px;
  padding-top: 14px;
  border-top: 1px dashed rgba(15, 23, 42, 0.12);
}

code {
  background: #f4f4f5;
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
