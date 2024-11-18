<template>
  <div class="pod-metrics">
    <h2 class="page-title">Pod 资源监控</h2>
    <div class="control-panel">
      <el-space wrap>
        <el-select v-model="sortBy" placeholder="排序方式">
          <el-option label="按 Pod IP 排序" value="name" />
          <el-option label="按 CPU 占用排序" value="cpu" />
          <el-option label="按内存占用排序" value="memory" />
        </el-select>
        
        <el-input
          v-model="searchHostIP"
          placeholder="搜索宿主机 IP"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        
        <el-input
          v-model="searchNamespace"
          placeholder="搜索命名空间"
          clearable
        >
          <template #prefix>
            <el-icon><Collection /></el-icon>
          </template>
        </el-input>
        
        <el-button type="primary" @click="searchPods">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </el-space>
    </div>

    <el-row :gutter="20">
      <el-col 
        v-for="pod in sortedPods" 
        :key="pod.hostname"
        :xs="24"
        :sm="12"
        :md="8"
        :lg="6"
      >
        <el-card 
          class="pod-card" 
          shadow="hover"
          @click="navigateToPodDetails(pod.hostname)"
        >
          <template #header>
            <div class="card-header">
              <el-tooltip 
                :content="pod.hostname"
                placement="top"
                :show-after="500"
              >
                <span class="pod-name">{{ pod.hostname }}</span>
              </el-tooltip>
            </div>
          </template>

          <div class="pod-info">
            <el-descriptions :column="1" size="small" border>
              <el-descriptions-item label="Pod IP">
                {{ pod.ip }}
              </el-descriptions-item>
              <el-descriptions-item label="节点 IP">
                {{ pod.node_ip }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <div class="metrics">
            <div class="metric-item">
              <span class="metric-label">
                <el-icon><CPU /></el-icon>
                CPU 使用率
              </span>
              <el-progress 
                :percentage="Number(pod.cpu_usage.toFixed(1))"
                :format="(val) => `${val}微核`"
                :color="[
                      { color: '#409EFF', percentage: 5000 },
                      { color: '#ed0202', percentage: 10000 }
                    ]"
              />
            </div>

            <div class="metric-item">
              <span class="metric-label">
                <el-icon><Monitor /></el-icon>
                内存使用率
              </span>
              <el-progress 
                :percentage="Number(pod.memory_usage.toFixed(1))"
                :format="(val) => `${val}MB`"
                :color="[
                      { color: '#409EFF', percentage: 5000 },
                      { color: '#ed0202', percentage: 10000 }
                    ]"
              />
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Collection, Cpu, Monitor } from '@element-plus/icons-vue'
import { api } from '../api'
const router = useRouter()

// 状态
const pods = ref([])
const sortBy = ref('name')
const searchHostIP = ref('')
const searchNamespace = ref('')

// 计算属性：排序后的Pod列表
const sortedPods = computed(() => {
  let filteredPods = pods.value

  // 应用搜索过滤
  if (searchHostIP.value) {
    filteredPods = filteredPods.filter(pod => 
      pod.node_ip.includes(searchHostIP.value)
    )
  }
  
  if (searchNamespace.value) {
    filteredPods = filteredPods.filter(pod => 
      pod.namespace?.includes(searchNamespace.value)
    )
  }

  // 应用排序
  return [...filteredPods].sort((a, b) => {
    switch (sortBy.value) {
      case 'cpu':
        return b.cpu_usage - a.cpu_usage
      case 'memory':
        return b.memory_usage - a.memory_usage
      default:
        const nodeA = a.node_ip || 'zzzzzzzzzzzzz'
        const nodeB = b.node_ip || 'zzzzzzzzzzzzz'
        
        const nodeIPCompare = nodeA.localeCompare(nodeB)
        if (nodeIPCompare !== 0) {
          return nodeIPCompare
        }
        return a.ip.localeCompare(b.ip)
    }
  })
})

// 方法
function getMetricStatus(value) {
  if (value >= 90) return 'exception'
  if (value >= 70) return 'warning'
  return 'success'
}

function navigateToPodDetails(hostname) {
  router.push(`/pod-details/${hostname}`)
}

async function loadPodData() {
  try {
    const params = new URLSearchParams({
      host_ip: searchHostIP.value,
      namespace: searchNamespace.value
    })
    
    const data = await api.getPodDashboard(`${params}`)
    pods.value = data
  } catch (error) {
    console.error('加载Pod数据失败:', error)
    ElMessage.error('加载Pod数据时出现错误')
  }
}

function searchPods() {
  loadPodData()
}

// 生命周期钩子
let timer
onMounted(() => {
  loadPodData()
  timer = setInterval(loadPodData, 2000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.pod-metrics {
  padding: 24px;
  background-color: #f9fafb;
  min-height: 100vh;
}

.control-panel {
  margin: 24px 0;
}

.pod-card {
  margin-bottom: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  background-color: #ffffff;
  transition: all 0.3s ease;
  border: 1px solid #f0f0f0;
}

.pod-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.08);
}

.card-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  color: #1f2937;
}

.pod-name {
  font-size: 15px;
  font-weight: 500;
}

.pod-info {
  margin: 16px 0;
}

.metrics {
  padding: 8px 0;
}

.metric-item {
  margin-bottom: 16px;
}

.metric-label {
  color: #6b7280;
  font-size: 14px;
  margin-bottom: 8px;
}

:deep(.el-progress-bar__outer) {
  background-color: #f3f4f6;
}

:deep(.el-progress-bar__inner) {
  background-color: #60a5fa;
}

:deep(.el-input__inner),
:deep(.el-select) {
  border-radius: 6px;
}

:deep(.el-button--primary) {
  background-color: #60a5fa;
  border-color: #60a5fa;
  border-radius: 6px;
}

:deep(.el-button--primary:hover) {
  background-color: #3b82f6;
  border-color: #3b82f6;
}

:deep(.el-descriptions) {
  padding: 8px;
}

:deep(.el-descriptions__label) {
  color: #6b7280;
}
</style> 