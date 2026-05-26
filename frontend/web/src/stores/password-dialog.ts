import { reactive } from 'vue'

interface PasswordDialogState {
  visible: boolean
  forced: boolean
}

const state = reactive<PasswordDialogState>({
  visible: false,
  forced: false,
})

export function usePasswordDialogState() {
  return state
}

export function openPasswordDialog(forced = false) {
  console.info('[password-dialog] open', { forced })
  state.visible = true
  state.forced = forced
}

export function closePasswordDialog() {
  console.info('[password-dialog] close', { forced: state.forced })
  state.visible = false
  state.forced = false
}

export function syncForcedPasswordDialog(active: boolean) {
  if (active) {
    openPasswordDialog(true)
    return
  }
  if (state.forced) {
    closePasswordDialog()
  }
}
