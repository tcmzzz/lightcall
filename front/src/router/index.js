import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
import NotFound from '../views/NotFound.vue'

const router = createRouter({
  //history: createWebHistory(import.meta.env.BASE_URL),
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue')
    },
    {
      path: '/',
      name: 'home',
      redirect: '/task/table',
      component: () => import('@/MainLayout.vue'),
      children: [
        {
          path: '/sysconfig',
          name: 'sysconfig',
          redirect: '/sysconfig/config',
          component: () => import('@/views/sys/SysIndex.vue'),
          children: [
            {
              path: 'config',
              component: () => import('@/views/sys/SysConfig.vue')
            },
            {
              path: 'user',
              component: () => import('@/views/sys/SysUser.vue')
            }
          ]
        },
        {
          path: '/call-config',
          name: 'call-config',
          redirect: '/call-config/number',
          component: () => import('@/views/call-config/CallConfigIndex.vue'),
          children: [
            {
              path: 'number',
              component: () => import('@/views/call-config/NumTable.vue')
            },
            {
              path: 'gateway',
              component: () => import('@/views/call-config/OutGateway.vue')
            }
          ]
        },
        {
          path: '/task',
          name: 'task',
          redirect: '/task/table',
          children: [
            {
              path: 'table',
              component: () => import('@/views/task/TaskTable.vue')
            }
          ]
        },
        {
          path: '/objective',
          name: 'objective',
          redirect: '/objective/table',
          component: () => import('@/views/objective/ObjectiveIndex.vue'),
          children: [
            {
              path: 'table',
              component: () => import('@/views/objective/ObjectiveTable.vue')
            },
            {
              path: 'detail/:id',
              component: () => import('@/views/objective/ObjectiveDetail.vue')
            }
          ]
        },
        { path: '/:pathMatch(.*)*', name: 'NotFound', component: NotFound }
      ]
    }
  ]
})

export default router
