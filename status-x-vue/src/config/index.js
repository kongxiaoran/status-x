// 获取当前环境
const env = import.meta.env.VITE_APP_ENV || 'development'
console.log('Current ENV:', env) // 用于调试

// 环境配置
const config = {
  development: {
    baseURL: 'http://localhost:12800',
    wsURL: 'ws://localhost:12800'
  },
  production: {
    baseURL: 'http://10.15.97.66:42800',
    wsURL: 'ws://10.15.97.66:42800'
  }
}

// 当前环境的配置
const currentConfig = config[env]
console.log('Current Config:', currentConfig) // 用于调试

export default {
  // 当前环境
  env,
  
  // API 基础URL
  baseURL: currentConfig.baseURL,
  
  // WebSocket URL
  wsURL: currentConfig.wsURL,
  
  // WebSocket 路径
  wsDashboardPath: '/ws/dashboard',
  
  // 获取完整的 WebSocket URL
  getDashboardWsURL() {
    return `${this.wsURL}${this.wsDashboardPath}`
  }
} 