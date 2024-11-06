<template>
  <div class="pod-details">
    <el-page-header 
      :content="`Pod 详情: ${podName}`"
      @back="router.back()"
    />
    
    <div class="time-range">
      <el-space wrap>
        <el-date-picker
          v-model="timeRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          :shortcuts="shortcuts"
          :default-time="defaultTime"
        />
        
        <el-button type="primary" @click="fetchMetrics">
          查询
        </el-button>
      </el-space>
    </div>

    <el-row :gutter="20">
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>CPU 使用率历史数据</span>
            </div>
          </template>
          <div ref="cpuChartRef" class="chart" />
        </el-card>
      </el-col>
      
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>内存使用率历史数据</span>
            </div>
          </template>
          <div ref="memoryChartRef" class="chart" />
        </el-card>
      </el-col>
      
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>网络使用率历史数据</span>
            </div>
          </template>
          <div ref="networkChartRef" class="chart" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const podName = route.params.pod

// 图表引用
const cpuChartRef = ref(null)
const memoryChartRef = ref(null)
const networkChartRef = ref(null)

// 图表实例
let cpuChart = null
let memoryChart = null
let networkChart = null

// 时间范围
const timeRange = ref([])
const defaultTime = [
  new Date(2000, 1, 1, 0, 0, 0),
  new Date(2000, 1, 1, 23, 59, 59),
]

// 快捷时间范围
const shortcuts = [
  {
    text: '最近30分钟',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 30 * 60 * 1000)
      return [start, end]
    },
  },
  {
    text: '最近1小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000)
      return [start, end]
    },
  },
  {
    text: '最近6小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 6 * 3600 * 1000)
      return [start, end]
    },
  },
]

// 初始化图表
function initCharts() {
  if (cpuChartRef.value) {
    cpuChart = echarts.init(cpuChartRef.value)
  }
  if (memoryChartRef.value) {
    memoryChart = echarts.init(memoryChartRef.value)
  }
  if (networkChartRef.value) {
    networkChart = echarts.init(networkChartRef.value)
  }
}

// 更新图表配置
function updateChart(chart, title, data, unit = '') {
  if (!chart) return

  const option = {
    title: {
      text: title,
      left: 'center'
    },
    tooltip: {
      trigger: 'axis',
      formatter: function(params) {
        const time = params[0].axisValue
        const value = params[0].value
        return `${time}<br/>${params[0].seriesName}: ${value}${unit}`
      }
    },
    xAxis: {
      type: 'category',
      data: data.timestamps,
      axisLabel: {
        rotate: 45
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: `{value}${unit}`
      }
    },
    series: [{
      name: title,
      type: 'line',
      smooth: true,
      data: data.values,
      areaStyle: {
        opacity: 0.3
      },
      lineStyle: {
        width: 2
      }
    }],
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true
    }
  }

  chart.setOption(option)
}

// 获取指标数据
async function fetchMetrics() {
  if (!timeRange.value?.length) {
    ElMessage.warning('请选择时间范围')
    return
  }

  try {
    const [start, end] = timeRange.value
    const params = new URLSearchParams({
      pod: podName,
      start: start.toISOString(),
      end: end.toISOString()
    })

    const response = await fetch(`/api/pod-metrics?${params}`)
    const data = await response.json()

    const metrics = {
      cpu: { timestamps: [], values: [] },
      memory: { timestamps: [], values: [] },
      network: { timestamps: [], values: [] }
    }

    data.forEach(entry => {
      const timestamp = new Date(entry._time).toLocaleString()
      
      switch (entry._field) {
        case 'cpu_usage':
          metrics.cpu.timestamps.push(timestamp)
          metrics.cpu.values.push(entry._value)
          break
        case 'memory_usage':
          metrics.memory.timestamps.push(timestamp)
          metrics.memory.values.push(entry._value)
          break
        case 'network_usage':
          metrics.network.timestamps.push(timestamp)
          metrics.network.values.push(entry._value)
          break
      }
    })

    updateChart(cpuChart, 'CPU 使用率', metrics.cpu, '微核')
    updateChart(memoryChart, '内存使用率', metrics.memory, 'MB')
    updateChart(networkChart, '网络使用率', metrics.network, 'MB/s')
  } catch (error) {
    console.error('获取指标数据失败:', error)
    ElMessage.error('获取指标数据时出现错误')
  }
}

// 窗口大小改变时重绘图表
function handleResize() {
  cpuChart?.resize()
  memoryChart?.resize()
  networkChart?.resize()
}

// 生命周期钩子
onMounted(() => {
  initCharts()
  window.addEventListener('resize', handleResize)
  
  // 设置默认时间范围为最近30分钟
  const end = new Date()
  const start = new Date()
  start.setTime(start.getTime() - 30 * 60 * 1000)
  timeRange.value = [start, end]
  
  fetchMetrics()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  cpuChart?.dispose()
  memoryChart?.dispose()
  networkChart?.dispose()
})
</script>

<style scoped>
.pod-details {
  padding: 24px;
  background-color: #f9fafb;
  min-height: 100vh;
}

.time-range {
  margin: 24px 0;
}

.chart-card {
  margin-bottom: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  background-color: #ffffff;
}

.chart {
  height: 400px;
  width: 100%;
  padding: 16px;
}

.card-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 500;
}

:deep(.el-date-editor) {
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
</style> 