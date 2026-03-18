import request from './request'

export function updateMember(data) {
  return request.post('/admin/user/update-member', data)
}
