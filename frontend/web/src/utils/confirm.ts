import { openConfirmDialog, type ConfirmDialogOptions } from '@/stores/confirm-dialog'

export type { ConfirmDialogOptions }

export function confirmDialog(message: string, title = '操作确认', options: ConfirmDialogOptions = {}) {
  return openConfirmDialog(message, title, options)
}
