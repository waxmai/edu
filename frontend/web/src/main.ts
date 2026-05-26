import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/index.css'
import { useAuthStore } from '@/stores/auth'
import { installChunkLoadRecovery } from '@/utils/runtime-recovery'

installChunkLoadRecovery()

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

const authStore = useAuthStore(pinia)
authStore.bootstrap().finally(() => {
  app.use(router)
  app.mount('#app')
})
