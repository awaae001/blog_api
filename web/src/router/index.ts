import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import Login from '@/views/Login.vue'
import Panel from '@/views/Panel.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/panel'
  },
  {
    path: '/panel/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/panel',
    name: 'Panel',
    component: Panel,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')

  if (to.meta.requiresAuth) {
    if (!token) {
      next('/panel/login')
    } else {
      next()
    }
  } else {
    if (token && to.path === '/panel/login') {
      next('/panel')
    } else {
      next()
    }
  }
})

export default router
