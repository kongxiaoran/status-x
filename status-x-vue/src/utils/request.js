import { ElMessage } from 'element-plus'
import config from '../config'

export async function request(url, options = {}) {
  try {
    const response = await fetch(`${config.baseURL}${url}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    const data = await response.json()
    return data
  } catch (error) {
    console.error('API request failed:', error)
    ElMessage.error('请求失败：' + error.message)
    throw error
  }
} 