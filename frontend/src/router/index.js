import { createRouter, createWebHistory } from 'vue-router'
import { ensureAccessToken } from '../api/request'

const routes = [
  {
    path: '/',
    name: 'Landing',
    component: () => import('../views/Landing.vue'),
    meta: { guest: true, allowAuthed: true },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { guest: true },
  },
  {
    path: '/share/:token',
    name: 'ShareDownload',
    component: () => import('../views/ShareDownload.vue'),
    meta: { guest: true },
  },
  {
    path: '/drive',
    component: () => import('../components/AppLayout.vue'),
    meta: { auth: true },
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('../views/Home.vue'),
      },
      {
        path: 'recycle',
        name: 'Recycle',
        component: () => import('../views/Recycle.vue'),
      },
      {
        path: 'admin',
        name: 'Admin',
        component: () => import('../views/Admin.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

let authInitialized = false

router.beforeEach(async (to, _from, next) => {
  if (!authInitialized || (to.meta.auth && !localStorage.getItem('accessToken'))) {
    await ensureAccessToken()
    authInitialized = true
  }

  const token = localStorage.getItem('accessToken')
  if (to.meta.auth && !token) {
    next('/login')
  } else if (to.meta.guest && token && !to.meta.allowAuthed && to.name !== 'ShareDownload') {
    next('/drive')
  } else {
    next()
  }
})

export default router
