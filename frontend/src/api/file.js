import request from './request'

export function getFileList(parentId = 0) {
  return request.get('/file/list', { params: { parent_id: parentId } })
}

export function uploadFile(file, parentId = 0, onProgress, signal) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('parent_id', parentId)
  return request.post('/file/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress,
    signal,
  })
}

export function createFolder(name, parentId = 0) {
  return request.post('/file/upload-folder', { name, parent_id: parentId })
}

export function deleteFile(id) {
  return request.delete(`/file/delete/${id}`)
}

export function permanentDeleteFile(id) {
  return request.delete(`/file/permanent-delete/${id}`)
}

export function getRecycleList() {
  return request.get('/file/recycle')
}

export function restoreFile(id) {
  return request.post(`/file/restore/${id}`)
}

export function downloadFile(id) {
  return request.get(`/file/download/${id}`)
}
