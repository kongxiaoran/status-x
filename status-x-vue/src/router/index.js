import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue')
  },
  {
    path: '/home',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue')
  },
  {
    path: '/alert-config',
    name: 'AlertConfig',
    component: () => import('../views/AlertConfig.vue')
  },
  {
    path: '/pod-metrics',
    name: 'PodMetrics',
    component: () => import('../views/PodMetrics.vue')
  },
  {
    path: '/host-manager',
    name: 'HostManager',
    component: () => import('../views/HostManager.vue')
  },
  {
    path: '/host-metrics/:host',
    name: 'HostMetrics',
    component: () => import('../views/HostMetrics.vue')
  },
  {
    path: '/pod-details/:pod',
    name: 'PodDetails',
    component: () => import('../views/PodDetails.vue')
  }
]

export default createRouter({
  history: createWebHistory('/vue/'),
  routes
}) 