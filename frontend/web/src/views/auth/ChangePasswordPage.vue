<template>
  <div class="change-password-page">
    <el-card class="change-password-card" shadow="hover">
      <div class="change-password-header">
        <h1>{{ pageTitle }}</h1>
        <p>{{ pageDescription }}</p>
      </div>

      <el-alert
        :title="alertTitle"
        :type="mustChangePassword ? 'warning' : 'info'"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      />

      <el-form label-position="top" :model="form" @submit.prevent>
        <el-form-item label="原密码">
          <el-input v-model="form.oldPassword" type="password" show-password placeholder="请输入当前密码" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="form.newPassword" type="password" show-password placeholder="请输入新密码，至少 6 位" />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input v-model="form.confirmPassword" type="password" show-password placeholder="请再次输入新密码" @keyup.enter="handleSubmit" />
        </el-form-item>

        <el-space fill style="width: 100%">
          <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ mustChangePassword ? '确认修改' : '保存新密码' }}</el-button>
          <el-button :disabled="submitting" @click="handleLogout">退出登录</el-button>
        </el-space>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import { changePassword } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const router = useRouter()
const authStore = useAuthStore()
const submitting = ref(false)
const mustChangePassword = computed(() => Boolean(authStore.currentUser?.mustChangePassword))
const pageTitle = computed(() => (mustChangePassword.value ? '首次登录请先修改密码' : '修改当前账号密码'))
const pageDescription = computed(() => (
  mustChangePassword.value
    ? '为了避免初始化账号长期可用，完成改密后才能继续访问业务页面。'
    : '修改当前登录账号的密码，修改成功后将继续保留当前登录状态。'
))
const alertTitle = computed(() => (
  mustChangePassword.value
    ? '当前账号仍处于首次登录状态'
    : '请妥善保管新密码，避免与旧密码或默认口令重复'
))

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

async function handleSubmit() {
  if (!form.oldPassword.trim() || !form.newPassword.trim() || !form.confirmPassword.trim()) {
    message.error('请完整填写密码信息')
    return
  }
  if (form.newPassword.trim().length < 6) {
    message.error('新密码至少 6 位')
    return
  }
  if (form.newPassword !== form.confirmPassword) {
    message.error('两次输入的新密码不一致')
    return
  }
  if (form.oldPassword === form.newPassword) {
    message.error('新密码不能与原密码相同')
    return
  }

  submitting.value = true
  try {
    const response = await changePassword({
      oldPassword: form.oldPassword,
      newPassword: form.newPassword,
    })
    const payload = response.data?.data
    if (!payload?.token || !payload?.refreshToken || !payload.userInfo) {
      throw new Error('改密响应缺少新会话信息')
    }
    authStore.setToken(payload.token)
    authStore.setRefreshToken(payload.refreshToken)
    authStore.setCurrentUser(payload.userInfo)
    message.success('密码修改成功')
    await router.replace('/dashboard')
  } catch (error: any) {
    console.error(error)
    message.error(extractErrorMessage(error, '修改密码失败，请确认原密码和新密码后重试'))
  } finally {
    submitting.value = false
  }
}

async function handleLogout() {
  await authStore.logout()
  await router.replace('/login')
}
</script>

<style scoped>
.change-password-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background:
    radial-gradient(circle at top, rgba(245, 158, 11, 0.16), transparent 26%),
    linear-gradient(135deg, #fff7e6 0%, #fff 100%);
}

.change-password-card {
  width: 560px;
  max-width: 100%;
  border-radius: 20px;
  overflow: hidden;
}

.change-password-header {
  margin-bottom: 20px;
}

.change-password-header h1 {
  margin: 0 0 8px;
  font-size: 30px;
  letter-spacing: -0.03em;
}

.change-password-header p {
  margin: 0;
  color: #606266;
}
</style>
