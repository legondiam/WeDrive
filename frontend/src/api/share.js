import request from './request'

export function createShare(data) {
  return request.post('/share/create', data)
}

export function downloadShare(data) {
  return request.post('/share/download', data)
}
