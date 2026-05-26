import { reactive, ref } from 'vue'
import { fetchLessonPackageList } from '@/api/lessonPackage'
import { usePlatformViewScope } from '@/composables/usePlatformViewScope'
import type { LessonPackageItem, LessonPackageSearchParams } from '@/modules/lesson-package/types/lesson-package'
import { normalizeListPayload } from '@/utils/api'
import { cleanQueryParams } from '@/utils/query'

export function useLessonPackageList() {
  const { withPlatformViewScope } = usePlatformViewScope()
  const loading = ref(false)
  const data = ref<LessonPackageItem[]>([])
  const total = ref(0)
  const query = reactive<LessonPackageSearchParams>({
    studentId: '',
    courseId: '',
    status: '',
    lowLessonAlert: false,
    paymentStatus: '',
    pageNum: 1,
    pageSize: 10,
  })

  async function load(extra?: Partial<LessonPackageSearchParams>) {
    loading.value = true
    try {
      const params = withPlatformViewScope({
        ...query,
        ...extra,
      })
      const response = await fetchLessonPackageList(cleanQueryParams(params))
      const normalized = normalizeListPayload<LessonPackageItem>(response.data?.data)
      data.value = normalized.list
      total.value = normalized.total
      Object.assign(query, params)
    } finally {
      loading.value = false
    }
  }

  function reset() {
    query.studentId = ''
    query.courseId = ''
    query.status = ''
    query.lowLessonAlert = false
    query.paymentStatus = ''
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
