import { createAppMessage } from '@/stores/app-message'
import type { AppMessageInput } from '@/types/message'

const DEFAULT_DURATION = 2600

function normalizeOptions(options: AppMessageInput) {
  if (typeof options === 'string') {
    return { message: options, duration: DEFAULT_DURATION }
  }
  return {
    message: options?.message || '',
    duration: options?.duration ?? DEFAULT_DURATION,
  }
}

function showMessage(type: 'success' | 'warning' | 'error' | 'info', options: AppMessageInput) {
  const normalized = normalizeOptions(options)
  return createAppMessage(type, normalized.message, normalized.duration)
}

export const message = {
  success(options: AppMessageInput) {
    return showMessage('success', options)
  },
  warning(options: AppMessageInput) {
    return showMessage('warning', options)
  },
  error(options: AppMessageInput) {
    return showMessage('error', options)
  },
  info(options: AppMessageInput) {
    return showMessage('info', options)
  },
}
