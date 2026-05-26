<template>
  <router-view />
  <AppMessageHost />
  <AppConfirmDialogHost />
  <AppPasswordDialogHost />
</template>

<script setup lang="ts">
import { onErrorCaptured } from 'vue'
import { useRoute } from 'vue-router'
import AppConfirmDialogHost from '@/components/AppConfirmDialogHost.vue'
import AppMessageHost from '@/components/AppMessageHost.vue'
import AppPasswordDialogHost from '@/components/AppPasswordDialogHost.vue'
import { message } from '@/utils/message'
import { saveRuntimeError, summarizeUnknownError } from '@/utils/runtime-error'

const route = useRoute()

onErrorCaptured((error) => {
  const detail = summarizeUnknownError(error)
  const routePath = route.fullPath || window.location.pathname || '/'
  saveRuntimeError({
    scope: 'vue',
    route: routePath,
    message: detail,
    time: new Date().toISOString(),
  })
  console.error('[runtime][vue]', routePath, error)
  message.error(`页面加载异常（${routePath}）：${detail}`)
  return false
})
</script>
