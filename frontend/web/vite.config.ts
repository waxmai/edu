import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'node:path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

function resolveElementPlusChunk(id: string): string | undefined {
  if (!id.includes('element-plus')) {
    return undefined
  }

  if (id.includes('/components/table') || id.includes('/components/table-v2')) {
    return 'vendor-el-table'
  }

  if (
    id.includes('/components/date-picker') ||
    id.includes('/components/time-picker') ||
    id.includes('/components/calendar') ||
    id.includes('/components/time-select')
  ) {
    return 'vendor-el-datetime'
  }

  return 'vendor-element-plus'
}

function resolveVendorChunk(id: string): string | undefined {
  if (!id.includes('node_modules')) {
    return undefined
  }

  if (id.includes('echarts')) {
    return 'vendor-echarts'
  }

  if (id.includes('axios')) {
    return 'vendor-axios'
  }

  if (id.includes('@element-plus/icons-vue')) {
    return 'vendor-element-plus-icons'
  }

  if (id.includes('@floating-ui')) {
    return 'vendor-floating-ui'
  }

  if (id.includes('async-validator')) {
    return 'vendor-async-validator'
  }

  if (id.includes('dayjs')) {
    return 'vendor-dayjs'
  }

  if (id.includes('lodash-unified')) {
    return 'vendor-lodash'
  }

  if (id.includes('@vueuse/core') || id.includes('@vueuse/shared')) {
    return 'vendor-vueuse'
  }

  if (id.includes('element-plus') || id.includes('@element-plus')) {
    return resolveElementPlusChunk(id)
  }

  if (id.includes('/vue-router/')) {
    return 'vendor-vue-router'
  }

  if (id.includes('/pinia/')) {
    return 'vendor-pinia'
  }

  return undefined
}

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'vue-router'],
      resolvers: [ElementPlusResolver()],
      dts: './src/auto-imports.d.ts',
      eslintrc: {
        enabled: false,
      },
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: './src/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    sourcemap: false,
    chunkSizeWarningLimit: 700,
    rollupOptions: {
      output: {
        manualChunks: resolveVendorChunk,
      },
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5174,
    headers: {
      'Cache-Control': 'no-store, no-cache, must-revalidate, proxy-revalidate',
      Pragma: 'no-cache',
      Expires: '0',
    },
  },
})
