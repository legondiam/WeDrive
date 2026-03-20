import axios from 'axios'
import { toast } from 'vue-sonner'
import router from '../router'

const service = axios.create({
  baseURL: '/api/v1',
  timeout: 120000,
  withCredentials: true,
})

service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('accessToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code !== 0) {
      if (res.code === 1005) {
        return tryRefreshToken(response.config)
      }
      toast.error(res.msg || '请求失败')
      return Promise.reject(new Error(res.msg || '请求失败'))
    }
    return res
  },
  (error) => {
    if (error.response?.status === 401 || error.response?.data?.code === 1005) {
      return tryRefreshToken(error.config)
    }
    toast.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

let isRefreshing = false
let pendingRequests = []

async function tryRefreshToken(failedConfig) {
  if (failedConfig._retried) {
    goLogin()
    return Promise.reject(new Error('登录已过期'))
  }

  if (isRefreshing) {
    return new Promise((resolve, reject) => {
      pendingRequests.push({ resolve, reject, config: failedConfig })
    })
  }

  isRefreshing = true
  try {
    const { data } = await axios.post('/api/v1/user/refresh', null, {
      withCredentials: true,
    })
    if (data.code === 0 && data.data?.accessToken) {
      localStorage.setItem('accessToken', data.data.accessToken)
      failedConfig._retried = true
      failedConfig.headers.Authorization = `Bearer ${data.data.accessToken}`

      pendingRequests.forEach(({ resolve, config }) => {
        config._retried = true
        config.headers.Authorization = `Bearer ${data.data.accessToken}`
        resolve(service(config))
      })
      pendingRequests = []
      return service(failedConfig)
    } else {
      goLogin()
      return Promise.reject(new Error('刷新令牌失败'))
    }
  } catch {
    goLogin()
    return Promise.reject(new Error('刷新令牌失败'))
  } finally {
    isRefreshing = false
  }
}

function goLogin() {
  localStorage.removeItem('accessToken')
  pendingRequests.forEach(({ reject }) => reject(new Error('登录已过期')))
  pendingRequests = []
  router.push('/login')
}

export default service
