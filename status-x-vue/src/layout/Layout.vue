<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside width="200px" class="sidebar">
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        router
      >
        <!-- 菜单内容 -->
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container class="main-container">
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-container {
  min-height: 100vh;
}

/* 固定侧边栏 */
.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  background-color: #304156;
  z-index: 1000;
  overflow-y: auto;
}

/* 侧边栏菜单样式 */
.sidebar-menu {
  height: 100%;
  border-right: none;
}

/* 主内容区域左边需要留出侧边栏的宽度 */
.main-container {
  margin-left: 200px; /* 与侧边栏宽度相同 */
  min-height: 100vh;
}

/* 设置滚动条样式 */
.sidebar::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-thumb {
  background: #475365;
  border-radius: 3px;
}

.sidebar::-webkit-scrollbar-track {
  background: #304156;
}
</style> 