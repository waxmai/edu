import { reactive, ref } from 'vue'
import { fetchScheduleList } from '@/api/schedule'
import { usePlatformViewScope } from '@/composables/usePlatformViewScope'
import { normalizeListPayload } from '@/utils/api'
import { cleanQueryParams } from '@/utils/query'
import type { ScheduleItem } from '@/modules/schedule/types/schedule'

export function useScheduleCalendar() {
  const { withPlatformViewScope } = usePlatformViewScope()
  const loading = ref(false)
  const data = ref<ScheduleItem[]>([])
  const total = ref(0)
  const query = reactive({
    studentId: '',
    teacherId: '',
    lessonPackageId: '',
    scheduleStatus: '',
    startDate: '',
    endDate: '',
    pageNum: 1,
    pageSize: 10,
  })

  async function load(extra?: Record<string, unknown>) {
    loading.value = true
    try {
      const params = withPlatformViewScope({
        ...query,
        ...extra,
      })
      const response = await fetchScheduleList(cleanQueryParams(params))
      const normalized = normalizeListPayload<ScheduleItem>(response.data?.data)
      data.value = normalized.list.map((item) => ({
        ...item,
        status: item.scheduleStatus || item.status || 'scheduled',
      }))
      total.value = normalized.total
      Object.assign(query, params)
    } finally {
      loading.value = false
    }
  }

  function reset() {
    query.studentId = ''
    query.teacherId = ''
    query.lessonPackageId = ''
    query.scheduleStatus = ''
    query.startDate = ''
    query.endDate = ''
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
