<template>
  <div class="host-metrics">
    <el-page-header 
      :content="`主机监控: ${hostIp}`"
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
      <!-- CPU使用率图表 -->
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>CPU 使用率历史数据</span>
              <el-tooltip content="CPU使用率超过90%将触发告警" placement="top">
                <el-icon><Warning /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <div ref="cpuChartRef" class="chart" />
        </el-card>
      </el-col>

      <!-- 内存使用率图表 -->
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>内存使用率历史数据</span>
              <el-tooltip content="内存使用率超过90%将触发告警" placement="top">
                <el-icon><Warning /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <div ref="memoryChartRef" class="chart" />
        </el-card>
      </el-col>

      <!-- 磁盘使用率图表 -->
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>磁盘使用率历史数据</span>
              <el-tooltip content="磁盘使用率超过85%将触发告警" placement="top">
                <el-icon><Warning /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <div ref="diskChartRef" class="chart" />
        </el-card>
      </el-col>

      <!-- 网络IO图表 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>网络IO历史数</span>
            </div>
          </template>
          <div ref="networkChartRef" class="chart" />
        </el-card>
      </el-col>

      <!-- 磁盘IO图表 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>磁盘IO历史数据</span>
            </div>
          </template>
          <div ref="diskIoChartRef" class="chart" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Warning } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { api } from '../api'

const route = useRoute()
const router = useRouter()
const hostIp = route.params.host

// 图表引用
const cpuChartRef = ref(null)
const memoryChartRef = ref(null)
const diskChartRef = ref(null)
const networkChartRef = ref(null)
const diskIoChartRef = ref(null)

// 图表实例
let charts = {
  cpu: null,
  memory: null,
  disk: null,
  network: null,
  diskIo: null
}

// 时间范围
const timeRange = ref([])
const defaultTime = [
  new Date(2000, 1, 1, 0, 0, 0),
  new Date(2000, 1, 1, 23, 59, 59),
]

// 快捷时范围
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
  const refs = {
    cpu: cpuChartRef.value,
    memory: memoryChartRef.value,
    disk: diskChartRef.value,
    network: networkChartRef.value,
    diskIo: diskIoChartRef.value
  }

  Object.entries(refs).forEach(([key, ref]) => {
    if (ref) {
      charts[key] = echarts.init(ref)
    }
  })
}

// 更新图表配置
function updateChart(chart, title, data, unit = '', areaStyle = true) {
  if (!chart) return

  // 确保数据按时间排序并处理数据缺失
  const sortedData = data.timestamps.map((time, index) => ({
    time,
    value: data.values[index]
  }))
  .sort((a, b) => a.time - b.time)
  .filter(item => item.value !== null) // 过滤掉空值

  // 处理数据断点，在数据缺失处断开线条
  const connectedData = []
  let currentSegment = []

  sortedData.forEach((point, index) => {
    if (index > 0) {
      const timeDiff = point.time - sortedData[index - 1].time
      // 如果时间间隔超过2分钟，认为是数据断点
      if (timeDiff > 120000) { // 2分钟 = 120000毫秒
        if (currentSegment.length > 0) {
          connectedData.push(currentSegment)
          currentSegment = []
        }
      }
    }
    currentSegment.push([point.time, point.value])
    
    // 处理最后一段数据
    if (index === sortedData.length - 1 && currentSegment.length > 0) {
      connectedData.push(currentSegment)
    }
  })

  const totalDuration = sortedData.length > 1 
    ? sortedData[sortedData.length - 1].time - sortedData[0].time 
    : 3600000
  const timeInterval = Math.max(Math.floor(totalDuration / 10), 1000)

  const option = {
    title: {
      text: title,
      left: 'center'
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
        animation: false,
        snap: false
      },
      formatter: function(params) {
        if (params[0].value === '-') {
          return '暂无数据'
        }

        const currentTime = Number(params[0].axisValue)
        
        // 找到最接近的数据点
        const point = sortedData.reduce((prev, curr) => {
          return Math.abs(curr.time - currentTime) < Math.abs(prev.time - currentTime) ? curr : prev
        })

        const date = new Date(point.time)
        const timeStr = date.toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit'
        })

        return `
          <div style="padding: 3px;">
            <div style="margin-bottom: 3px;">${timeStr}</div>
            <div style="display: flex; justify-content: space-between;">
              <span style="margin-right: 15px;">${params[0].seriesName}:</span>
              <span style="font-weight: bold;">${point.value.toFixed(2)}${unit}</span>
            </div>
          </div>
        `
      }
    },
    xAxis: {
      type: 'time',
      minInterval: 1000,
      maxInterval: timeInterval,
      splitNumber: 10,
      axisPointer: {
        snap: false,
        label: {
          show: true,
          formatter: (params) => {
            const date = new Date(params.value)
            return date.toLocaleTimeString('zh-CN', {
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit'
            })
          }
        }
      },
      axisLabel: {
        formatter: (value) => {
          const date = new Date(value)
          return date.toLocaleTimeString('zh-CN', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
          })
        },
        rotate: 45
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: `{value}${unit}`
      },
      min: 0,
      max: unit === '%' ? 100 : undefined
    },
    series: connectedData.map(segment => ({
      name: title,
      type: 'line',
      smooth: true,
      symbol: 'none',
      sampling: 'lttb',
      data: segment,
      connectNulls: false, // 不连接空值点
      areaStyle: areaStyle ? {
        opacity: 0.3
      } : undefined,
      lineStyle: {
        width: 2
      },
      emphasis: {
        focus: 'series',
        lineStyle: {
          width: 3
        }
      }
    })),
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true
    },
    dataZoom: [
      {
        type: 'inside',
        start: 0,
        end: 100
      },
      {
        type: 'slider',
        start: 0,
        end: 100
      }
    ]
  }

  chart.setOption(option)
}

// 处理指标数据 - 使用原始数据点而不是固定间隔
function processMetricsData(data) {
  // 按指标类型分组数据
  const metrics = {
    cpu: { timestamps: [], values: [] },
    memory: { timestamps: [], values: [] },
    disk: { timestamps: [], values: [] },
    network: { timestamps: [], values: [] },
    diskIo: { timestamps: [], values: [] }
  }

  // 直接使用原始数据点
  data.forEach(entry => {
    const timestamp = new Date(entry._time).getTime()
    
    switch (entry._field) {
      case 'cpu_usage':
        metrics.cpu.timestamps.push(timestamp)
        metrics.cpu.values.push(entry._value)
        break
      case 'memory_usage':
        metrics.memory.timestamps.push(timestamp)
        metrics.memory.values.push(entry._value)
        break
      case 'disk_usage':
        metrics.disk.timestamps.push(timestamp)
        metrics.disk.values.push(entry._value)
        break
      case 'network_io':
        metrics.network.timestamps.push(timestamp)
        metrics.network.values.push(entry._value)
        break
      case 'read_write_io':
        metrics.diskIo.timestamps.push(timestamp)
        metrics.diskIo.values.push(entry._value)
        break
    }
  })

  return metrics
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
      host: hostIp,
      start: start.toISOString(),
      end: end.toISOString()
    })

    const data = await api.getHostMetrics(`${params}`)
    const metrics = processMetricsData(data)

    updateChart(charts.cpu, 'CPU 使用率', metrics.cpu, '%')
    updateChart(charts.memory, '内存使用率', metrics.memory, '%')
    updateChart(charts.disk, '磁盘使用率', metrics.disk, '%')
    updateChart(charts.network, '网络IO', metrics.network, 'MB/s', false)
    updateChart(charts.diskIo, '磁盘IO', metrics.diskIo, 'MB/s', false)
  } catch (error) {
    console.error('获取指标数据失败:', error)
    ElMessage.error('获取指标数据时出现错误')
  }
}

// 窗口大小改变时重绘图表
function handleResize() {
  Object.values(charts).forEach(chart => {
    chart?.resize()
  })
}

// 生命周期钩子
onMounted(() => {
  initCharts()
  window.addEventListener('resize', handleResize)
  
  // 设置默认时间范围为最近30钟
  const end = new Date()
  const start = new Date()
  start.setTime(start.getTime() - 30 * 60 * 1000)
  timeRange.value = [start, end]
  
  fetchMetrics()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  Object.values(charts).forEach(chart => {
    chart?.dispose()
  })
})
</script>

<style scoped>
.host-metrics {
  padding: 20px;
}

.time-range {
  margin: 20px 0;
}

.chart-card {
  margin-bottom: 20px;
}

.chart {
  height: 400px;
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.el-icon {
  font-size: 18px;
  color: #E6A23C;
  cursor: help;
}
</style> 