import request from './request'

export function login(data) {
  return request.post('/user/login', data)
}

export function register(data) {
  return request.post('/user/register', data)
}

export function refreshToken() {
  return request.post('/user/refresh')
}

export function getUserInfo() {
  return request.get('/user/info')
}
