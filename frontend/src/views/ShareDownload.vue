<template>
  <div class="share-page">
    <div class="share-card">
      <div class="share-header">
        <el-icon :size="48" color="#409eff"><Cloudy /></el-icon>
        <h1>WeDrive 文件分享</h1>
      </div>

      <template v-if="!downloadReady">
        <el-form size="large" @keyup.enter="handleGetFile">
          <el-form-item>
            <el-input
              v-model="key"
              placeholder="请输入提取码（无密码则留空）"
              maxlength="4"
            >
              <template #prefix>
                <el-icon><Key /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              class="share-btn"
              :loading="loading"
              @click="handleGetFile"
            >
              获取文件
            </el-button>
          </el-form-item>
        </el-form>
      </template>

      <template v-else>
        <div class="file-ready">
          <el-icon :size="56" color="#67c23a"><CircleCheckFilled /></el-icon>
          <h2>{{ fileName }}</h2>
          <p>文件已就绪，点击下方按钮下载</p>
          <el-button type="primary" size="large" class="share-btn" @click="doDownload">
            <el-icon><Download /></el-icon>
            下载文件
          </el-button>
        </div>
      </template>

      <div class="share-footer">
        <router-link to="/login">登录 WeDrive</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { downloadShare } from '../api/share'

const route = useRoute()
const key = ref('')
const loading = ref(false)
const downloadReady = ref(false)
const downloadUrl = ref('')
const fileName = ref('')

async function handleGetFile() {
  const shareToken = route.params.token
  if (!shareToken) return

  loading.value = true
  try {
    const payload = { share_token: shareToken }
    if (key.value.trim()) {
      payload.key = key.value.trim()
    }
    const res = await downloadShare(payload)
    downloadUrl.value = res.data.URL
    fileName.value = res.data.FileName
    downloadReady.value = true
  } catch {
    /* handled by interceptor */
  } finally {
    loading.value = false
  }
}

function doDownload() {
  if (!downloadUrl.value) return
  const a = document.createElement('a')
  a.href = downloadUrl.value
  a.download = fileName.value
  a.target = '_blank'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}
</script>

<style scoped>
.share-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  padding: 20px;
}

.share-card {
  background: #fff;
  border-radius: 16px;
  padding: 48px 40px 36px;
  width: 100%;
  max-width: 440px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.share-header {
  margin-bottom: 32px;
}

.share-header h1 {
  margin-top: 12px;
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.share-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  border-radius: 8px;
}

.file-ready {
  padding: 16px 0;
}

.file-ready h2 {
  margin: 16px 0 8px;
  font-size: 18px;
  word-break: break-all;
}

.file-ready p {
  color: #909399;
  font-size: 14px;
  margin-bottom: 24px;
}

.share-footer {
  margin-top: 24px;
  font-size: 13px;
  color: #909399;
}
</style>
