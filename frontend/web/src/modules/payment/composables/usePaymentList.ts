import { reactive, ref } from 'vue'
import { fetchPaymentList } from '@/api/payment'
import { usePlatformViewScope } from '@/composables/usePlatformViewScope'
import type { PaymentItem, PaymentSearchParams } from '@/modules/payment/types/payment'
import { normalizeListPayload } from '@/utils/api'
import { cleanQueryParams } from '@/utils/query'

export function usePaymentList() {
  const { withPlatformViewScope } = usePlatformViewScope()
  const loading = ref(false)
  const data = ref<PaymentItem[]>([])
  const total = ref(0)
  const query = reactive<PaymentSearchParams>({
    studentId: '',
    lessonPackageId: '',
    paymentType: '',
    paymentStatus: '',
    paymentMethod: '',
    startDate: '',
    endDate: '',
    pageNum: 1,
    pageSize: 10,
  })

  async function load(extra?: Partial<PaymentSearchParams>) {
    loading.value = true
    try {
      const params = withPlatformViewScope({
        ...query,
        ...extra,
      })
      const response = await fetchPaymentList(cleanQueryParams(params))
      const normalized = normalizeListPayload<PaymentItem>(response.data?.data)
      data.value = normalized.list
      total.value = normalized.total
      Object.assign(query, params)
    } finally {
      loading.value = false
    }
  }

  function reset() {
    query.studentId = ''
    query.lessonPackageId = ''
    query.paymentType = ''
    query.paymentStatus = ''
    query.paymentMethod = ''
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
