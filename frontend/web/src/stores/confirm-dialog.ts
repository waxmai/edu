import { reactive } from 'vue'

export type ConfirmDialogType = 'warning' | 'info' | 'success' | 'error'

export interface ConfirmDialogOptions {
  confirmButtonText?: string
  cancelButtonText?: string
  type?: ConfirmDialogType
}

interface ConfirmDialogState {
  visible: boolean
  title: string
  message: string
  type: ConfirmDialogType
  confirmButtonText: string
  cancelButtonText: string
  resolver: null | ((value: 'confirm') => void)
  rejecter: null | ((reason?: any) => void)
}

const state = reactive<ConfirmDialogState>({
  visible: false,
  title: '操作确认',
  message: '',
  type: 'warning',
  confirmButtonText: '确认',
  cancelButtonText: '取消',
  resolver: null,
  rejecter: null,
})

function reset() {
  state.visible = false
  state.message = ''
  state.title = '操作确认'
  state.type = 'warning'
  state.confirmButtonText = '确认'
  state.cancelButtonText = '取消'
  state.resolver = null
  state.rejecter = null
}

export function useConfirmDialogState() {
  return state
}

export function openConfirmDialog(message: string, title = '操作确认', options: ConfirmDialogOptions = {}) {
  return new Promise<'confirm'>((resolve, reject) => {
    state.visible = true
    state.message = message
    state.title = title
    state.type = options.type || 'warning'
    state.confirmButtonText = options.confirmButtonText || '确认'
    state.cancelButtonText = options.cancelButtonText || '取消'
    state.resolver = resolve
    state.rejecter = reject
  })
}

export function confirmConfirmDialog() {
  const resolve = state.resolver
  reset()
  resolve?.('confirm')
}

export function cancelConfirmDialog(reason: any = 'cancel') {
  const reject = state.rejecter
  reset()
  reject?.(reason)
}
