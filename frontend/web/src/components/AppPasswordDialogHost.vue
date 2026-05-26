<template>
  <el-dialog
    :model-value="visible"
    :title="pageTitle"
    width="520px"
    append-to-body
    align-center
    :close-on-click-modal="!forced"
    :close-on-press-escape="!forced"
    :show-close="!forced"
    class="app-password-dialog"
    @close="handleCancel"
  >
    <div class="app-password-dialog__intro">
      <el-alert
        :title="alertTitle"
        :type="forced ? 'warning' : 'info'"
        :closable="false"
        show-icon
      />
    </div>

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
    </el-form>

    <template #footer>
      <div class="app-password-dialog__footer">
        <el-button v-if="forced" :disabled="submitting" @click="handleLogout">退出登录</el-button>
        <el-button v-else :disabled="submitting" @click="handleCancel">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ forced ? '确认修改' : '保存新密码' }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { changePassword } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { closePasswordDialog, usePasswordDialogState } from '@/stores/password-dialog'
import { extractErrorMessage } from '@/utils/error'
import { message } from '@/utils/message'

const router = useRouter()
const authStore = useAuthStore()
const passwordState = usePasswordDialogState()
const submitting = ref(false)

const visible = computed(() => authStore.forcePasswordChangeActive || passwordState.visible)
const forced = computed(() => authStore.forcePasswordChangeActive)

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const pageTitle = computed(() => (forced.value ? '首次登录请先修改密码' : '修改当前账号密码'))
const alertTitle = computed(() => (
  forced.value
    ? '当前账号仍处于首次登录状态，完成改密后才能继续操作'
    : '请妥善保管新密码，避免与旧密码或默认口令重复'
))

watch(
  () => visible.value,
  (visible) => {
    if (!visible) {
      form.oldPassword = ''
      form.newPassword = ''
      form.confirmPassword = ''
    }
  },
)

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
    if (!payload?.token || !payload?.refreshToken || !payload?.userInfo) {
      throw new Error('改密响应缺少新会话信息')
    }
    authStore.setToken(payload.token)
    authStore.setRefreshToken(payload.refreshToken)
    authStore.setCurrentUser(payload.userInfo)
    authStore.deactivateForcePasswordChange()
    closePasswordDialog()
    message.success('密码修改成功')
    if (router.currentRoute.value.path === '/change-password') {
      await router.replace('/dashboard')
    }
  } catch (error: any) {
    console.error(error)
    message.error(extractErrorMessage(error, '修改密码失败，请确认原密码和新密码后重试'))
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  if (forced.value) {
    return
  }
  closePasswordDialog()
}

async function handleLogout() {
  await authStore.logout()
  closePasswordDialog()
  await router.replace('/login')
}
</script>
