import type { Component } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { routes, handleHotUpdate } from 'vue-router/auto-routes'

declare module 'vue-router' {
  interface RouteMeta {
    /** 无需登录即可访问 */
    public?: boolean
    /** 侧边栏菜单标题，有值才会出现在菜单里 */
    title?: string
    /** 菜单图标 */
    icon?: Component
    /** 菜单排序，越小越靠前 */
    order?: number
    /** 菜单固定在底部 */
    bottom?: boolean
  }
}

const router = createRouter({
  history: createWebHistory('/panel/'),
  routes: [...routes, { path: '/:pathMatch(.*)*', redirect: '/' }]
})

router.beforeEach((to) => {
  const loggedIn = !!localStorage.getItem('token')
  if (to.meta.public) {
    return loggedIn && to.path === '/login' ? '/' : true
  }
  return loggedIn || '/login'
})

if (import.meta.hot) {
  handleHotUpdate(router)
}

export default router
