import { request } from '../utils/request'

export const api = {
  // 仪表盘
  getDashboard() {
    return request('/api/dashboard')
  },

  // Pod相关
  getPodDashboard(params) {
    const query = new URLSearchParams(params).toString()
    return request(`/api/pod-dashboard?${query}`)
  },

  getPodMetrics(params) {
    const query = new URLSearchParams(params).toString()
    return request(`/api/pod-metrics?${query}`)
  },

  // 主机相关
  getHostMetrics(params) {
    const query = new URLSearchParams(params).toString()
    return request(`/api/host-metrics?${query}`)
  },

  // 主机管理
  getHosts() {
    return request('/api/host-management')
  },

  addHost(data) {
    return request('/api/host-management', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  updateHost(data) {
    return request('/api/host-management', {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  },

  deleteHost(ip) {
    return request(`/api/host-management?ip=${ip}`, {
      method: 'DELETE'
    })
  },

  // 告警配置
  getAlertConfig() {
    return request('/api/alert-metrics')
  },

  updateAlertConfig(data) {
    return request('/api/alert-config', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  }
} 