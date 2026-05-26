import { reactive } from 'vue'

export type AppMessageType = 'success' | 'warning' | 'error' | 'info'

export interface AppMessageItem {
  id: number
  type: AppMessageType
  text: string
  visible: boolean
}

const DEFAULT_DURATION = 2600
const state = reactive({
  items: [] as AppMessageItem[],
})

let seed = 1
const timers = new Map<number, number>()

function remove(id: number) {
  const timer = timers.get(id)
  if (timer) {
    clearTimeout(timer)
    timers.delete(id)
  }
  const index = state.items.findIndex((item) => item.id === id)
  if (index >= 0) {
    state.items.splice(index, 1)
  }
}

function dismiss(id: number) {
  const item = state.items.find((entry) => entry.id === id)
  if (!item) {
    return
  }
  item.visible = false
  window.setTimeout(() => remove(id), 180)
}

function push(type: AppMessageType, text: string, duration = DEFAULT_DURATION) {
  if (!text.trim()) {
    return { close: () => undefined }
  }

  const same = state.items.find((item) => item.text === text && item.type === type)
  if (same) {
    const timer = timers.get(same.id)
    if (timer) {
      clearTimeout(timer)
    }
    same.visible = true
    timers.set(same.id, window.setTimeout(() => dismiss(same.id), duration))
    return { close: () => dismiss(same.id) }
  }

  const id = seed++
  state.items.push({ id, type, text: text.trim(), visible: true })
  timers.set(id, window.setTimeout(() => dismiss(id), duration))
  return { close: () => dismiss(id) }
}

export function useAppMessageState() {
  return state
}

export function createAppMessage(type: AppMessageType, text: string, duration?: number) {
  return push(type, text, duration)
}

export function dismissAppMessage(id: number) {
  dismiss(id)
}
