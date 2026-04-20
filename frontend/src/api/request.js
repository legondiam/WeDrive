import axios from 'axios'
import { toast } from 'vue-sonner'
import router from '../router'

function createRequestError(message, code) {
  const err = new Error(message || '请求失败')
  if (typeof code !== 'undefined') {
    err.code = code
  }
  return err
}

function shouldToastBusinessError(code) {
  return code !== 3003 && code !== 3008 && code !== 3009
}

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
      if (shouldToastBusinessError(res.code)) {
        toast.error(res.msg || '请求失败')
      }
      return Promise.reject(createRequestError(res.msg || '请求失败', res.code))
    }
    return res
  },
  (error) => {
    if (error.response?.status === 401 || error.response?.data?.code === 1005) {
      return tryRefreshToken(error.config)
    }
    const code = error.response?.data?.code
    const message = error.response?.data?.msg || error.message || '网络错误'
    if (shouldToastBusinessError(code)) {
      toast.error(message)
    }
    return Promise.reject(createRequestError(message, code))
  }
)

let refreshPromise = null

export async function refreshAccessToken() {
  if (!refreshPromise) {
    refreshPromise = axios
      .post('/api/v1/user/refresh', null, {
        withCredentials: true,
      })
      .then(({ data }) => {
        if (data.code === 0 && data.data?.accessToken) {
          localStorage.setItem('accessToken', data.data.accessToken)
          return data.data.accessToken
        }
        throw new Error(data.msg || '刷新令牌失败')
      })
      .finally(() => {
        refreshPromise = null
      })
  }

  return refreshPromise
}

export async function ensureAccessToken() {
  const token = localStorage.getItem('accessToken')
  if (token) {
    return token
  }

  try {
    return await refreshAccessToken()
  } catch {
    return null
  }
}

async function tryRefreshToken(failedConfig) {
  if (failedConfig._retried) {
    goLogin()
    return Promise.reject(new Error('登录已过期'))
  }

  try {
    const newToken = await refreshAccessToken()
    failedConfig._retried = true
    failedConfig.headers = failedConfig.headers || {}
    failedConfig.headers.Authorization = `Bearer ${newToken}`
    return service(failedConfig)
  } catch {
    goLogin()
    return Promise.reject(new Error('刷新令牌失败'))
  }
}

function goLogin() {
  localStorage.removeItem('accessToken')
  if (router.currentRoute.value.path !== '/login') {
    router.push('/login')
  }
}

export default service
