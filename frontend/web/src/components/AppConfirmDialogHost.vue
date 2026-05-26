<template>
  <el-dialog
    :model-value="confirmState.visible"
    :title="confirmState.title"
    width="420px"
    append-to-body
    align-center
    :close-on-click-modal="false"
    :close-on-press-escape="true"
    :show-close="false"
    class="app-confirm-dialog"
    @close="cancelConfirmDialog()"
  >
    <div class="app-confirm-dialog__body">
      <span class="app-confirm-dialog__icon" :class="`app-confirm-dialog__icon--${confirmState.type}`">
        {{ iconMap[confirmState.type] }}
      </span>
      <div class="app-confirm-dialog__content">{{ confirmState.message }}</div>
    </div>

    <template #footer>
      <div class="app-confirm-dialog__footer">
        <el-button @click="cancelConfirmDialog()">{{ confirmState.cancelButtonText }}</el-button>
        <el-button :type="confirmButtonType" @click="confirmConfirmDialog()">{{ confirmState.confirmButtonText }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { cancelConfirmDialog, confirmConfirmDialog, useConfirmDialogState } from '@/stores/confirm-dialog'

const confirmState = useConfirmDialogState()

const iconMap = {
  warning: '!',
  error: '✕',
  success: '✓',
  info: 'i',
} as const

const confirmButtonType = computed(() => (confirmState.type === 'error' ? 'danger' : 'primary'))
</script>
