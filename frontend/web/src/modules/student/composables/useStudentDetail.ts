import { ref } from 'vue'
import { fetchStudentDetail } from '@/api/student'
import type { StudentDetail } from '@/modules/student/types/student'

export function useStudentDetail() {
  const loading = ref(false)
  const detail = ref<StudentDetail | null>(null)

  async function load(id: number | string) {
    loading.value = true
    try {
      const response = await fetchStudentDetail(id)
      detail.value = response.data?.data || null
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    detail,
    load,
  }
}
