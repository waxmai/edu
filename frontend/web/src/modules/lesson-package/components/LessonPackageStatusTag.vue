<template>
  <el-tag :type="tagType" effect="light" round class="status-tag">{{ label }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
}>()

const label = computed(() => {
  switch (props.status) {
    case 'pending':
      return '未生效'
    case 'active':
      return '生效中'
    case 'exhausted':
      return '已耗尽'
    case 'expired':
      return '已过期'
    case 'closed':
      return '已关闭'
    default:
      return props.status || '-'
  }
})

const tagType = computed(() => {
  switch (props.status) {
    case 'active':
      return 'success'
    case 'pending':
      return 'warning'
    case 'exhausted':
      return 'danger'
    case 'expired':
      return 'info'
    case 'closed':
      return 'info'
    default:
      return 'info'
  }
})
</script>
