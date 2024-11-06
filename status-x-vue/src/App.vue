<template>
  <div class="app-container">
    <SideBar class="app-sidebar" />
    <div class="main-container">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import SideBar from './components/SideBar.vue'
</script>

<style>
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
}

.app-container {
  min-height: 100vh;
}

.main-container {
  margin-left: 200px; /* 与侧边栏宽度相同 */
  min-height: 100vh;
  background-color: #f4f4f4;
  padding: 20px;
  box-sizing: border-box;
}
</style> 