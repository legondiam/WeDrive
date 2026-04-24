import request from './request'
import axios from 'axios'

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

export function instantUpload(payload) {
  return request.post('/file/instant-upload', payload)
}

export function prepareInstantUpload(payload) {
  return request.post('/file/instant-upload/prepare', payload)
}

export function quickCheck(payload) {
  return request.post('/file/quick-check', payload)
}

export function initChunkUpload(payload) {
  return request.post('/file/upload/init', payload)
}

export function signPartUpload(payload) {
  return request.post('/file/upload/sign-part', payload)
}

export function reportUploadedPart(payload) {
  return request.post('/file/upload/report-part', payload)
}

export function uploadChunkDirect(uploadUrl, chunk, headers = {}, onProgress, signal) {
  return axios.put(uploadUrl, chunk, {
    headers,
    onUploadProgress: onProgress,
    signal,
  })
}

export function completeChunkUpload(payload) {
  return request.post('/file/upload/complete', payload)
}

export function createFolder(name, parentId = 0) {
  return request.post('/file/upload-folder', { name, parent_id: parentId })
}

export function deleteFile(id) {
  return request.delete(`/file/delete/${id}`)
}

export function batchDeleteFiles(ids) {
  return request.post('/file/batch-delete', { ids })
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
