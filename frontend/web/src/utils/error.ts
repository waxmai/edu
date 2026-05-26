function translateKnownErrorMessage(message: string) {
  if (message.includes('chk_lesson_package_paid_not_exceed_total')) {
    return '已缴金额不能超过总金额'
  }
  return message
}

export function extractErrorMessage(error: any, fallback = '操作失败，请稍后重试') {
  const responseData = error?.response?.data
  const candidates = [
    responseData?.message,
    responseData?.msg,
    responseData?.error,
    responseData?.businessCodeMsg,
    error?.message,
  ]

  for (const candidate of candidates) {
    if (typeof candidate === 'string' && candidate.trim()) {
      return translateKnownErrorMessage(candidate.trim())
    }
  }

  return fallback
}

export function isCancelError(error: unknown) {
  return error === 'cancel' || (error as any)?.message === 'cancel'
}
