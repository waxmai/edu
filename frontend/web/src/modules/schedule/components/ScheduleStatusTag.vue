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
    case 'scheduled':
      return '待上课'
    case 'completed':
      return '已完成'
    case 'leave':
      return '已请假'
    case 'rescheduled':
      return '已调课'
    case 'cancelled':
      return '已取消'
    case 'makeup_pending':
      return '待补课'
    default:
      return props.status || '-'
  }
})

const tagType = computed(() => {
  switch (props.status) {
    case 'scheduled':
      return 'warning'
    case 'completed':
      return 'success'
    case 'leave':
      return 'info'
    case 'rescheduled':
      return 'primary'
    case 'cancelled':
      return 'danger'
    case 'makeup_pending':
      return 'warning'
    default:
      return 'info'
  }
})
</script>
