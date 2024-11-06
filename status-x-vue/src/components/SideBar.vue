<template>
  <el-menu
    class="sidebar"
    :router="true"
    background-color="#333"
    text-color="#fff"
    active-text-color="#ffd04b"
  >
    <el-menu-item index="/home" style="margin-top: 30px;">
      <el-icon><DataLine /></el-icon>
      <span>仪表板</span>
    </el-menu-item>
    <el-menu-item index="/alert-config">
      <el-icon><Bell /></el-icon>
      <span>警报配置</span>
    </el-menu-item>
    <el-menu-item index="/pod-metrics">
      <el-icon><Monitor /></el-icon>
      <span>Pod监测</span>
    </el-menu-item>
    <el-menu-item index="/host-manager">
      <el-icon><Setting /></el-icon>
      <span>主机管理</span>
    </el-menu-item>
  </el-menu>
</template>

<script setup>
import { DataLine, Bell, Monitor, Setting } from '@element-plus/icons-vue'
</script>

<style scoped>
.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  width: 200px;
  height: 100vh;
  background-color: #333;
  z-index: 1000;
  overflow-y: auto;
}

.sidebar::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 3px;
}

.sidebar::-webkit-scrollbar-track {
  background: #333;
}

:deep(.el-menu-item) {
  border-left: 3px solid transparent;
}

:deep(.el-menu-item.is-active) {
  border-left: 3px solid #ffd04b;
  background-color: #252525 !important;
}

:deep(.el-menu-item:hover) {
  background-color: #252525 !important;
}

:deep(.el-menu-item .el-icon) {
  margin-right: 12px;
}
</style> 