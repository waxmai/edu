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
    case 'paid':
      return '已缴'
    case 'pending':
      return '待缴'
    case 'partial':
      return '部分已缴'
    case 'refunded':
      return '已退款'
    default:
      return props.status || '-'
  }
})

const tagType = computed(() => {
  switch (props.status) {
    case 'paid':
      return 'success'
    case 'pending':
      return 'warning'
    case 'partial':
      return 'danger'
    case 'refunded':
      return 'info'
    default:
      return 'info'
  }
})
</script>
