import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  base: '/vue/', // 设置根路径
  plugins: [vue()],
  build: {
    outDir: '../monitor-server/frontend',
  },
  define: {
    'import.meta.env.MODE': JSON.stringify(process.env.NODE_ENV || 'development')
  }
})
