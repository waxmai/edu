import { reactive, ref } from 'vue'
import { fetchStudentList } from '@/api/student'
import { usePlatformViewScope } from '@/composables/usePlatformViewScope'
import type { StudentItem, StudentSearchParams } from '@/modules/student/types/student'
import { normalizeListPayload } from '@/utils/api'
import { cleanQueryParams } from '@/utils/query'

export function useStudentList() {
  const { withPlatformViewScope } = usePlatformViewScope()
  const loading = ref(false)
  const data = ref<StudentItem[]>([])
  const total = ref(0)
  const query = reactive<StudentSearchParams>({
    studentName: '',
    subject: '',
    status: '',
    parentPhone: '',
    inactiveAlert: false,
    pageNum: 1,
    pageSize: 10,
  })

  async function load(extra?: Partial<StudentSearchParams>) {
    loading.value = true
    try {
      const params = withPlatformViewScope({
        ...query,
        ...extra,
      })
      const response = await fetchStudentList(cleanQueryParams(params))
      const payload = response.data?.data
      const normalized = normalizeListPayload<StudentItem>(payload)
      data.value = normalized.list
      total.value = normalized.total
      Object.assign(query, params)
    } finally {
      loading.value = false
    }
  }

  function reset() {
    query.studentName = ''
    query.subject = ''
    query.status = ''
    query.parentPhone = ''
    query.inactiveAlert = false
    query.pageNum = 1
    query.pageSize = 10
  }

  return {
    loading,
    data,
    total,
    query,
    load,
    reset,
  }
}
