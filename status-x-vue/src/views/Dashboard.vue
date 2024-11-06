<template>
  <div class="dashboard">
    <!-- <el-page-header content="中台服务器监控" :show-back="false" /> -->
    <h2 class="page-title">中台服务器监控</h2>
    
    <!-- 添加加载状态 -->
    <el-loading 
      v-if="store.loading" 
      :fullscreen="true" 
      text="加载中..."
    />

    <!-- 错误提示 -->
    <el-alert
      v-if="store.error"
      :title="store.error"
      type="error"
      show-icon
      closable
      @close="store.error = null"
    />

    <!-- 总览数据卡片 -->
    <el-row v-if="!store.error" :gutter="20" class="overview-cards">
      <el-col :xs="24" :sm="12" :md="4">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>总主机数</span>
              <el-icon><Monitor /></el-icon>
            </div>
          </template>
          <div class="card-value">
            {{ statistics.totalHosts }}
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="4">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>CPU总核心数</span>
              <el-icon><Cpu /></el-icon>
            </div>
          </template>
          <div class="card-value">
            {{ statistics.totalCpuCores }}
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="4">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>内存总量</span>
              <el-icon><Connection /></el-icon>
            </div>
          </template>
          <div class="card-value">
            {{ formatStorage(statistics.totalMemory) }}
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="4">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>磁盘总量</span>
              <el-icon><DataLine /></el-icon>
            </div>
          </template>
          <div class="card-value">
            {{ formatStorage(statistics.totalDisk) }}
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="24" :md="8">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>整体资源使用率</span>
              <el-icon><TrendCharts /></el-icon>
            </div>
          </template>
          <div class="usage-metrics">
            <div class="usage-item">
              <span class="usage-label">CPU</span>
              <span :class="['usage-value', getUsageClass(statistics.avgCpuUsage)]">
                {{ statistics.avgCpuUsage.toFixed(1) }}%
              </span>
            </div>
            <div class="usage-item">
              <span class="usage-label">内存</span>
              <span :class="['usage-value', getUsageClass(statistics.avgMemoryUsage)]">
                {{ statistics.avgMemoryUsage.toFixed(2) }}%
              </span>
            </div>
            <div class="usage-item">
              <span class="usage-label">磁盘</span>
              <span :class="['usage-value', getUsageClass(statistics.avgDiskUsage)]">
                {{ statistics.avgDiskUsage.toFixed(1) }}%
              </span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 控制面板 -->
    <div class="control-panel">
      <el-space wrap :size="30">
        <div class="control-item">
          <span>排序方式:</span>
          <el-select v-model="sortBy" placeholder="排序方式" style="min-width: 150px;">
            <el-option label="按标签分组" value="label" />
            <el-option label="按IP排序" value="ip" />
            <el-option label="按CPU占用排序" value="cpu" />
            <el-option label="按内存占用排序" value="memory" />
            <el-option label="按磁盘占用排序" value="disk" />
          </el-select>
        </div>
        
        <div class="control-item">
          <span>分组:</span>
          <el-select 
            v-model="selectedGroup" 
            placeholder="选择分组"
            clearable
            @clear="selectedGroup = ''"
            style="min-width: 250px;"
          >
            <el-option label="全部" value="" />
            <el-option 
              v-for="group in availableGroups" 
              :key="group" 
              :label="group" 
              :value="group" 
            />
          </el-select>
        </div>
        
        <div class="control-item">
          <span>IP筛选:</span>
          <el-input
            v-model="ipFilter"
            placeholder="输入IP进行筛选"
            clearable
            style="min-width: 200px;"
          />
        </div>
        
        <el-switch
          v-model="showDetail"
          active-text="详细"
          inactive-text="简略"
        />
      </el-space>
    </div>

    <!-- 主机列表 -->
    <div 
      v-for="(group, label, index) in groupedHosts" 
      :key="label" 
      class="host-group"
      :class="`group-theme-${index % 4}`"
    >
      <div v-if="sortBy === 'label'" class="group-header">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标签">
            {{ label }}
          </el-descriptions-item>
          <el-descriptions-item label="主机数量">
            {{ group.length }}
          </el-descriptions-item>
          <el-descriptions-item label="资源统计">
            <el-space wrap>
              <el-tag>核心数: {{ getGroupStats(group).totalCores }}</el-tag>
              <el-tag type="success">内存: {{ getGroupStats(group).totalMemSize }} GB</el-tag>
              <el-tag type="warning">磁盘: {{ getGroupStats(group).totalDiskSize }} GB</el-tag>
            </el-space>
          </el-descriptions-item>
          <el-descriptions-item label="平均使用率">
            <el-space wrap>
              <el-tag :type="getTagType(getGroupStats(group).avgCpu)">
                CPU: {{ getGroupStats(group).avgCpu }}%
              </el-tag>
              <el-tag :type="getTagType(getGroupStats(group).avgMemory)">
                内存: {{ getGroupStats(group).avgMemory }}%
              </el-tag>
              <el-tag :type="getTagType(getGroupStats(group).avgDisk)">
                磁盘: {{ getGroupStats(group).avgDisk }}%
              </el-tag>
            </el-space>
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <el-row :gutter="20">
        <el-col 
          v-for="host in group" 
          :key="host.ip" 
          :xs="24" 
          :sm="12" 
          :md="8" 
          :lg="6"
        >
          <el-card 
            class="host-card" 
            shadow="hover"
            @click="navigateToDetail(host.ip)"
          >
            <template #header>
              <div class="card-header">
                <el-text class="host-ip" truncated>
                  {{ host.ip }}
                </el-text>
                <el-tag size="small">{{ host.label || '未分组' }}</el-tag>
              </div>
            </template>

            <div class="host-info">
              <el-descriptions :column="2" border size="small">
                <el-descriptions-item label="CPU核心">
                  {{ host.cpu_cores }}
                </el-descriptions-item>
                <el-descriptions-item label="内存">
                  {{ (host.total_memory / 1024 / 1024 / 1024).toFixed(2) }} GB
                </el-descriptions-item>
                <el-descriptions-item label="磁盘">
                  {{ (host.total_disk / 1024 / 1024 / 1024).toFixed(2) }} GB
                </el-descriptions-item>
                <el-descriptions-item label="状态">
                  <el-tag :type="host.status !== 'online' ? 'success' : 'danger'">
                    {{ host.status !== 'online' ? '在线' : '离线' }}
                  </el-tag>
                </el-descriptions-item>
              </el-descriptions>
            </div>

            <div class="metrics">
              <div class="metric-item">
                <span class="metric-label">CPU使用率</span>
                <el-progress 
                  :percentage="Number(host.cpu_usage.toFixed(1))"
                  :color="getProgressColors('cpu')"
                />
              </div>

              <div class="metric-item">
                <span class="metric-label">内存使用率</span>
                <el-progress 
                  :percentage="Number(host.memory_usage.toFixed(1))"
                  :color="getProgressColors('memory')"
                />
              </div>

              <div class="metric-item">
                <span class="metric-label">磁盘使用率</span>
                <el-progress 
                  :percentage="Number(host.disk_usage.toFixed(1))"
                  :color="getProgressColors('disk')"
                />
              </div>

              <template v-if="showDetail">
                <div class="metric-item">
                  <span class="metric-label">网络IO</span>
                  <el-progress 
                    :percentage="Number((host.network_io * 10).toFixed(1))"
                    :format="() => `${host.network_io.toFixed(2)} MB`"
                    :color="[
                      { color: '#409EFF', percentage: 0 },
                      { color: '#67C23A', percentage: 100 }
                    ]"
                  />
                </div>

                <div class="metric-item">
                  <span class="metric-label">磁盘IO</span>
                  <el-progress 
                    :percentage="Number(host.read_write_io.toFixed(1))"
                    :format="() => `${host.read_write_io.toFixed(2)} MB`"
                    :color="[
                      { color: '#409EFF', percentage: 50 },
                      { color: '#67C23A', percentage: 100 }
                    ]"
                  />
                </div>

                <div class="metric-item">
                  <span class="metric-label">网络连接数</span>
                  <el-progress 
                    :percentage="Number((host.net_conn_count * 0.1).toFixed(1))"
                    :format="() => host.net_conn_count.toString()"
                    :color="[
                      { color: '#409EFF', percentage: 50 },
                      { color: '#67C23A', percentage: 100 }
                    ]"
                  />
                </div>
              </template>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMetricsStore } from '../stores/metrics'
import { Monitor, Cpu, Connection, DataLine } from '@element-plus/icons-vue'
import config from '../config'

const router = useRouter()
const route = useRoute()
const store = useMetricsStore()
let ws = null

// 状态
const sortBy = ref('label')
const showDetail = ref(localStorage.getItem('detailMode') === 'true')

// 新增状态
const selectedGroup = ref('')
const ipFilter = ref('')

// 监听 URL 参数变化
watch(
  () => route.query.ip,
  (newIp) => {
    ipFilter.value = newIp || ''
  },
  { immediate: true }
)

// 监听 ipFilter 变化，更新 URL
watch(ipFilter, (newValue) => {
  router.replace({
    query: {
      ...route.query,
      ip: newValue || undefined
    }
  })
})

// 计算属性
const statistics = computed(() => {
  const hosts = store.hosts
  const total = hosts.length
  
  if (total === 0) {
    return {
      totalHosts: 0,
      totalCpuCores: 0,
      totalMemory: 0,
      totalDisk: 0,
      avgCpuUsage: 0,
      avgMemoryUsage: 0,
      avgDiskUsage: 0
    }
  }

  const totals = hosts.reduce((acc, host) => ({
    cpuCores: acc.cpuCores + host.cpu_cores,
    memory: acc.memory + (host.total_memory / 1024 / 1024 / 1024), // 转换为GB
    disk: acc.disk + (host.total_disk / 1024 / 1024 / 1024), // 转换为GB
    usedCpu: acc.usedCpu + (host.cpu_cores * host.cpu_usage / 100), // CPU核心数 * 使用率
    usedMemory: acc.usedMemory + (host.total_memory * host.memory_usage / 100), // 总内存 * 使用率
    usedDisk: acc.usedDisk + (host.total_disk * host.disk_usage / 100), // 总磁盘 * 使用率
  }), {
    cpuCores: 0,
    memory: 0,
    disk: 0,
    usedCpu: 0,
    usedMemory: 0,
    usedDisk: 0
  })

  console.log(totals.memory)
  return {
    totalHosts: total,
    totalCpuCores: totals.cpuCores,
    totalMemory: totals.memory.toFixed(2),
    totalDisk: totals.disk.toFixed(2),
    avgCpuUsage: (totals.usedCpu / totals.cpuCores) * 100,
    avgMemoryUsage: (totals.usedMemory / (totals.memory * 1024 * 1024 * 1024)) * 100,
    avgDiskUsage: (totals.usedDisk / (totals.disk * 1024 * 1024 * 1024)) * 100
  }

})

const REFRESH_INTERVAL = 1000  // 1秒

const debouncedHosts = computed(() => {
  let hosts = store.hosts.map(host => ({
    ...host,
    cpu_usage: Number(host.cpu_usage.toFixed(1)),
    memory_usage: Number(host.memory_usage.toFixed(1)),
    disk_usage: Number(host.disk_usage.toFixed(1)),
    network_io: Number(host.network_io.toFixed(2)),
    read_write_io: Number(host.read_write_io.toFixed(2))
  }))

  // 应用 IP 筛选
  if (ipFilter.value) {
    hosts = hosts.filter(host => 
      host.ip.toLowerCase().includes(ipFilter.value.toLowerCase())
    )
  }

  return hosts
})

// 新增计算属性
const availableGroups = computed(() => {
  const groups = store.hosts.map(host => host.label || '未分组')
  return [...new Set(groups)].sort() // 对分组名进行排序
})

const groupedHosts = computed(() => {
  const hosts = debouncedHosts.value
  
  if (sortBy.value === 'label') {
    const grouped = groupByLabel(hosts)
    
    // 如果选择了特定分组，只返回该分组
    if (selectedGroup.value) {
      return {
        [selectedGroup.value]: grouped[selectedGroup.value] || []
      }
    }
    
    // 返回排序后的分组
    return Object.keys(grouped)
      .sort()
      .reduce((acc, key) => {
        acc[key] = grouped[key]
        return acc
      }, {})
  } else {
    // 如果选择了特定分组，只显示该分组的主机
    const filteredHosts = selectedGroup.value
      ? hosts.filter(host => (host.label || '未分组') === selectedGroup.value)
      : hosts
    return { '所有主机': sortHosts(filteredHosts) }
  }
})

// 方法
function groupByLabel(hosts) {
  const grouped = hosts.reduce((acc, host) => {
    const label = host.label || '未分组'
    if (!acc[label]) acc[label] = []
    acc[label].push(host)
    return acc
  }, {})

  // 对每个分组内的主机按 IP 排序
  Object.keys(grouped).forEach(label => {
    grouped[label].sort((a, b) => a.ip.localeCompare(b.ip))
  })

  return grouped
}

function sortHosts(hosts) {
  return [...hosts].sort((a, b) => {
    switch (sortBy.value) {
      case 'cpu':
        return b.cpu_usage - a.cpu_usage
      case 'memory':
        return b.memory_usage - a.memory_usage
      case 'disk':
        return b.disk_usage - a.disk_usage
      case 'ip':
      default:
        return a.ip.localeCompare(b.ip)
    }
  })
}

function getGroupStats(hosts) {
  const total = hosts.reduce((acc, host) => ({
    cpuUsage: acc.cpuUsage + host.cpu_usage,
    memoryUsage: acc.memoryUsage + host.memory_usage,
    diskUsage: acc.diskUsage + host.disk_usage,
    cpuCores: acc.cpuCores + host.cpu_cores,
    totalMemory: acc.totalMemory + host.total_memory,
    totalDisk: acc.totalDisk + host.total_disk
  }), {
    cpuUsage: 0, memoryUsage: 0, diskUsage: 0,
    cpuCores: 0, totalMemory: 0, totalDisk: 0
  })

  const count = hosts.length
  
  return {
    totalCores: total.cpuCores,
    totalMemSize: (total.totalMemory / 1024 / 1024 / 1024).toFixed(2),
    totalDiskSize: (total.totalDisk / 1024 / 1024 / 1024).toFixed(2),
    avgCpu: (total.cpuUsage / count).toFixed(2),
    avgMemory: (total.memoryUsage / count).toFixed(2),
    avgDisk: (total.diskUsage / count).toFixed(2)
  }
}

function getProgressColors(type) {
  switch (type) {
    case 'disk':
      return [
        { color: '#67C23A', percentage: 50 },   // 绿色 0-70%
        { color: '#E6A23C', percentage: 70 },  // 黄色 70-85%
        { color: '#F56C6C', percentage: 85 }   // 红色 85-100%
      ]
    case 'network':
      return [
          { color: '#67C23A', percentage: 50 },   // 绿色 0-70%
          { color: '#E6A23C', percentage: 70 },  // 黄色 70-85%
          { color: '#F56C6C', percentage: 85 }   // 红色 85-100%
        ]
    case 'diskio':
      return [
        { color: '#67C23A', percentage: 50 },   // 绿色 0-70%
        { color: '#E6A23C', percentage: 70 },  // 黄色 70-85%
        { color: '#F56C6C', percentage: 85 }   // 红色 85-100%
      ]
    case 'netconn':
      return [
        { color: '#67C23A', percentage: 50 },   // 绿色 0-70%
        { color: '#E6A23C', percentage: 70 },  // 黄色 70-85%
        { color: '#F56C6C', percentage: 85 }   // 红色 85-100%
      ]
    default: // CPU 和内存使用相同的阈值
      return [
        { color: '#67C23A', percentage: 50 },   // 绿色 0-70%
        { color: '#E6A23C', percentage: 70 },  // 黄色 70-90%
        { color: '#F56C6C', percentage: 90 }   // 红色 90-100%
      ]
  }
}

function getTagType(value) {
  if (value >= 90) return 'danger'
  if (value >= 70) return 'warning'
  return 'success'
}

function navigateToDetail(ip) {
  router.push(`/host-metrics/${ip}`)
}

// 替换原有的轮询逻辑为WebSocket连接
function connectWebSocket() {
  const wsUrl = config.getDashboardWsURL()
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('WebSocket connected')
    store.loading = false
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      store.hosts = data.hosts
      store.loading = data.loading
      store.error = data.error
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
      store.error = '数据解析错误'
    }
  }

  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
    store.error = 'WebSocket连接错误'
    store.loading = false
  }

  ws.onclose = () => {
    console.log('WebSocket disconnected')
    store.error = 'WebSocket连接已断开'
    store.loading = false
    
    // 尝试重新连接
    setTimeout(() => {
      connectWebSocket()
    }, 3000)
  }
}

// 生命周期钩子
onMounted(() => {
  store.loading = true
  connectWebSocket()
  
  if (route.query.ip) {
    ipFilter.value = route.query.ip
  }
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
})

// 添加格式化存储大小的方法
function formatStorage(sizeInGB) {
  if (sizeInGB >= 1024) {
    return `${(sizeInGB / 1024).toFixed(2)} TB`
  }
  return `${sizeInGB} GB`
}

// 修改使用率样式判断方法
function getUsageClass(value) {
  if (value >= 85) return 'danger'
  if (value >= 60) return 'warning'
  return 'success'
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.overview-cards {
  margin: 20px 0;
  row-gap: 20px;
}

.usage-card {
  margin-top: 20px;
}

.usage-metrics {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
}

.usage-item {
  text-align: center;
}

.usage-label {
  display: block;
  font-size: 14px;
  color: #606266;
  margin-bottom: 4px;
}

.usage-value {
  font-size: 20px;
  font-weight: bold;
}

.usage-value.success {
  color: #67C23A;
}

.usage-value.warning {
  color: #E6A23C;
}

.usage-value.danger {
  color: #F56C6C;
}

.card-value {
  font-size: 24px;
  font-weight: bold;
  text-align: center;
  color: #409EFF;
  padding: 10px 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  color: #303133;
}

.control-panel {
  margin: 20px 0;
}

.control-item {
  display: flex;
  align-items: center;
  gap: 8px;  /* 标签和控件之间的间距 */
}

.control-item span {
  white-space: nowrap;
}

.host-group {
  margin-bottom: 30px;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.group-theme-0 {
  background: linear-gradient(135deg, rgba(240, 248, 255, 0.8) 0%, rgba(229, 241, 255, 0.8) 100%);
  border: 1px solid rgba(64, 158, 255, 0.1);
}

.group-theme-1 {
  background: linear-gradient(135deg, rgba(245, 255, 245, 0.8) 0%, rgba(230, 245, 230, 0.8) 100%);
  border: 1px solid rgba(103, 194, 58, 0.1);
}

.group-theme-2 {
  background: linear-gradient(135deg, rgba(255, 248, 240, 0.8) 0%, rgba(255, 241, 229, 0.8) 100%);
  border: 1px solid rgba(230, 162, 60, 0.1);
}

.group-theme-3 {
  background: linear-gradient(135deg, rgba(255, 242, 242, 0.8) 0%, rgba(255, 235, 235, 0.8) 100%);
  border: 1px solid rgba(245, 108, 108, 0.1);
}

.group-header {
  margin-bottom: 25px;
}

.group-header .el-descriptions {
  background: rgba(255, 255, 255, 0.8);
  border-radius: 6px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.host-card {
  margin-bottom: 20px;
  transition: all 0.3s ease;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(0, 0, 0, 0.05);
}

.host-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.host-ip {
  font-weight: 600;
  color: #303133;
}

.host-info {
  margin-bottom: 15px;
}

.metrics {
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 10px 0;
}

.metric-item {
  background: rgba(255, 255, 255, 0.5);
  border-radius: 6px;
  
  .metric-label {
    display: block;
    margin-bottom: 8px;
    font-size: 14px;
    color: #606266;
    font-weight: 500;
  }
}

.host-info .el-descriptions {
  background: rgba(255, 255, 255, 0.5);
  border-radius: 6px;
  margin: 10px 0;
}

.el-progress-bar__outer {
  border-radius: 8px;
  background-color: rgba(0, 0, 0, 0.04);
}

.el-progress-bar__inner {
  border-radius: 8px;
  transition: width 0.5s ease-in-out;
}

:deep(.el-descriptions__label.el-descriptions__cell.is-bordered-label) {
  border-radius: 4px;
  background-color: rgba(235, 235, 235, 0.2) !important;
}

</style> 