<template>
  <div class="app-message-stack" aria-live="polite" aria-atomic="true">
    <transition-group name="app-message-fade">
      <div
        v-for="item in messageState.items"
        :key="item.id"
        class="app-message-item"
        :class="[`app-message-item--${item.type}`, { 'is-visible': item.visible }]"
        role="alert"
      >
        <span class="app-message-item__icon">{{ iconMap[item.type] }}</span>
        <span class="app-message-item__text">{{ item.text }}</span>
        <button class="app-message-item__close" type="button" aria-label="关闭提示" @click="dismissAppMessage(item.id)">×</button>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { dismissAppMessage, useAppMessageState } from '@/stores/app-message'

const messageState = useAppMessageState()

const iconMap = {
  success: '✓',
  warning: '!',
  error: '✕',
  info: 'i',
} as const
</script>
