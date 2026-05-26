import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
  state: () => ({
    sidebarCollapsed: false,
    globalLoading: false,
  }),
  actions: {
    setSidebarCollapsed(value: boolean) {
      this.sidebarCollapsed = value
    },
    setGlobalLoading(value: boolean) {
      this.globalLoading = value
    },
  },
})
