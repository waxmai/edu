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
    case 'active':
      return '在读'
    case 'suspended':
      return '停课'
    case 'graduated':
      return '结课'
    case 'inactive':
      return '停用'
    default:
      return props.status || '-'
  }
})

const tagType = computed(() => {
  switch (props.status) {
    case 'active':
      return 'success'
    case 'suspended':
      return 'warning'
    case 'graduated':
      return 'info'
    case 'inactive':
      return 'danger'
    default:
      return 'info'
  }
})
</script>
