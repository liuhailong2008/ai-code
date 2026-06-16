import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/login', name: 'login', component: () => import('@/pages/Login.vue') },
  { path: '/cas/callback', name: 'casCallback', component: () => import('@/pages/CasCallback.vue') },
  { path: '/',              redirect: '/dashboard' },
  { path: '/dashboard', name: 'dashboard', component: () => import('@/pages/Dashboard.vue') },
  { path: '/monitor',   name: 'monitor',   component: () => import('@/pages/DashboardMonitor.vue') },
  { path: '/leaf-connectivity', name: 'leaf-connectivity', component: () => import('@/pages/LeafConnectivity.vue') },
  { path: '/latency', name: 'latency', component: () => import('@/pages/LatencyMonitor.vue') },
  { path: '/topology',  name: 'topology',  component: () => import('@/pages/Topology.vue')  },
  { path: '/traffic',   name: 'traffic',   component: () => import('@/pages/Traffic.vue')   },
  { path: '/alerts',    name: 'alerts',    component: () => import('@/pages/Alerts.vue')    },
  { path: '/devices',   name: 'devices',   component: () => import('@/pages/Devices.vue')   },
  { path: '/reports',   name: 'reports',   component: () => import('@/pages/Reports.vue')   },
  { path: '/metrics-manage', name: 'metrics-manage', component: () => import('@/pages/MetricsManage.vue') },
  { path: '/metric-query', name: 'metric-query', component: () => import('@/pages/MetricQuery.vue') },
  { path: '/public-monitor', name: 'public-monitor', component: () => import('@/pages/PublicMonitor.vue') },
  { path: '/settings',  name: 'settings',  component: () => import('@/pages/Settings.vue')  },
  { path: '/settings-env', name: 'settings-env', component: () => import('@/pages/SettingsEnv.vue') },
  { path: '/statistics', name: 'statistics', component: () => import('@/pages/Statistics.vue') },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
  linkActiveClass: 'active',
})

router.beforeEach((to, from) => {
  const token = localStorage.getItem('pingmesh_token')
  const legacyLoggedIn = localStorage.getItem('pingmesh_logged_in') === '1'
  const isLoggedIn = !!token || legacyLoggedIn

  if (to.name !== 'login' && to.name !== 'casCallback' && to.name !== 'public-monitor' && !isLoggedIn) {
    return { name: 'login' }
  }

  if ((to.name === 'login' || to.name === 'casCallback') && isLoggedIn) {
    return { name: 'monitor' }
  }
})

export default router
