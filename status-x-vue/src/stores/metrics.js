import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api'

export const useMetricsStore = defineStore('metrics', () => {
  const hosts = ref([])
  const pods = ref([])
  const alertConfig = ref(null)
  const loading = ref(false)
  const error = ref(null)

  async function fetchHosts() {
    try {
      loading.value = true
      error.value = null
      const data = await api.getDashboard()
      hosts.value = data
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  async function fetchPods() {
    try {
      loading.value = true
      error.value = null
      const data = await api.getPodDashboard()
      pods.value = data
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  async function fetchAlertConfig() {
    try {
      loading.value = true
      error.value = null
      const data = await api.getAlertConfig()
      alertConfig.value = data
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  return {
    hosts,
    pods,
    alertConfig,
    loading,
    error,
    fetchHosts,
    fetchPods,
    fetchAlertConfig
  }
}) 