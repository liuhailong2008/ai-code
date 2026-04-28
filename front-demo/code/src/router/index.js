import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'IDCConnectivity',
    component: () => import('../views/MonitorCenter.vue'),
    meta: { title: '机房连通性' }
  },
  {
    path: '/leaf-connectivity',
    name: 'LeafConnectivity',
    component: () => import('../views/LeafConnectivity.vue'),
    meta: { title: 'Leaf连通性' }
  },
  {
    path: '/latency',
    name: 'LatencyMonitor',
    component: () => import('../views/LatencyMonitor.vue'),
    meta: { title: '机房耗时监控' }
  },
  {
    path: '/alerts/active',
    name: 'ActiveAlerts',
    component: () => import('../views/ActiveAlerts.vue'),
    meta: { title: '未恢复告警' }
  },
  {
    path: '/alerts/history',
    name: 'AlertHistory',
    component: () => import('../views/AlertHistory.vue'),
    meta: { title: '历史告警' }
  },
  {
    path: '/alerts/config',
    name: 'AlertConfig',
    component: () => import('../views/AlertConfig.vue'),
    meta: { title: '告警配置' }
  },
  {
    path: '/devices',
    name: 'DeviceManagement',
    component: () => import('../views/DeviceManagement.vue'),
    meta: { title: '设备管理' }
  },
  {
    path: '/strategies',
    name: 'StrategyManagement',
    component: () => import('../views/StrategyManagement.vue'),
    meta: { title: '策略管理' }
  },
  {
    path: '/tasks',
    name: 'TaskManagement',
    component: () => import('../views/TaskManagement.vue'),
    meta: { title: '任务管理' }
  },
  {
    path: '/agents',
    name: 'AgentManagement',
    component: () => import('../views/AgentManagement.vue'),
    meta: { title: '探测代理管理' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} - PingMesh` : 'PingMesh'
})

export default router
