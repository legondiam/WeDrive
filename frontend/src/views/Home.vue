<template>
  <div class="home">
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" @click="showUploadDialog = true">
          <el-icon><Upload /></el-icon>
          上传文件
        </el-button>
        <el-button @click="showFolderDialog = true">
          <el-icon><FolderAdd /></el-icon>
          新建文件夹
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-button :icon="Refresh" circle @click="fileStore.refresh()" />
      </div>
    </div>

    <!-- Breadcrumb -->
    <div class="breadcrumb-bar">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item
          v-for="(item, index) in fileStore.breadcrumbs"
          :key="item.id"
        >
          <span
            class="breadcrumb-link"
            :class="{ active: index === fileStore.breadcrumbs.length - 1 }"
            @click="fileStore.navigateTo(index)"
          >
            {{ item.name }}
          </span>
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <!-- File Table -->
    <el-table
      v-loading="fileStore.loading"
      :data="fileStore.files"
      class="file-table"
      empty-text="暂无文件"
      @row-dblclick="handleRowDblClick"
    >
      <el-table-column label="名称" min-width="300">
        <template #default="{ row }">
          <div class="file-name-cell">
            <el-icon :size="22" :color="row.is_folder ? '#e6a23c' : '#409eff'">
              <Folder v-if="row.is_folder" />
              <Document v-else />
            </el-icon>
            <span class="file-name">{{ row.file_name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="大小" width="120" align="center">
        <template #default="{ row }">
          <span class="file-meta">{{ row.is_folder ? '-' : row.file_size }}</span>
        </template>
      </el-table-column>
      <el-table-column label="修改时间" width="180" align="center">
        <template #default="{ row }">
          <span class="file-meta">{{ formatTime(row.updated_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" align="center" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="!row.is_folder"
            link
            type="primary"
            size="small"
            @click.stop="handleDownload(row)"
          >
            <el-icon><Download /></el-icon>
            下载
          </el-button>
          <el-button
            v-if="!row.is_folder"
            link
            type="success"
            size="small"
            @click.stop="handleShare(row)"
          >
            <el-icon><Share /></el-icon>
            分享
          </el-button>
          <el-button
            link
            type="danger"
            size="small"
            @click.stop="handleDelete(row)"
          >
            <el-icon><Delete /></el-icon>
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Upload Dialog -->
    <el-dialog
      v-model="showUploadDialog"
      title="上传文件"
      width="500px"
      destroy-on-close
    >
      <el-upload
        ref="uploadRef"
        drag
        :auto-upload="false"
        :on-change="handleFileChange"
        :file-list="uploadFileList"
        :limit="1"
        :on-exceed="() => ElMessage.warning('每次只能上传一个文件')"
      >
        <el-icon class="el-icon--upload" :size="48"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          拖拽文件到此处，或 <em>点击选择</em>
        </div>
      </el-upload>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">
          {{ uploading ? `上传中 ${uploadProgress}%` : '开始上传' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- New Folder Dialog -->
    <el-dialog
      v-model="showFolderDialog"
      title="新建文件夹"
      width="400px"
      destroy-on-close
    >
      <el-input
        v-model="newFolderName"
        placeholder="请输入文件夹名称"
        @keyup.enter="handleCreateFolder"
      />
      <template #footer>
        <el-button @click="showFolderDialog = false">取消</el-button>
        <el-button type="primary" :loading="creatingFolder" @click="handleCreateFolder">
          创建
        </el-button>
      </template>
    </el-dialog>

    <!-- Share Dialog -->
    <el-dialog
      v-model="showShareDialog"
      title="创建分享链接"
      width="480px"
      destroy-on-close
    >
      <template v-if="!shareResult">
        <el-form label-width="80px">
          <el-form-item label="文件">
            <span>{{ shareTarget?.file_name }}</span>
          </el-form-item>
          <el-form-item label="有效期">
            <el-radio-group v-model="shareForm.expiretime">
              <el-radio-button value="1">1天</el-radio-button>
              <el-radio-button value="7">7天</el-radio-button>
              <el-radio-button value="30">30天</el-radio-button>
              <el-radio-button value="permanent">永久</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="提取码">
            <el-input
              v-model="shareForm.key"
              placeholder="留空则无需提取码"
              maxlength="4"
              show-word-limit
              style="width: 200px"
            />
          </el-form-item>
        </el-form>
      </template>
      <template v-else>
        <div class="share-result">
          <el-alert type="success" :closable="false" show-icon>
            <template #title>分享链接已创建</template>
          </el-alert>
          <div class="share-link-box">
            <el-input :model-value="shareLink" readonly>
              <template #append>
                <el-button @click="copyShareLink">复制</el-button>
              </template>
            </el-input>
            <p v-if="shareForm.key" class="share-key">
              提取码：<strong>{{ shareForm.key }}</strong>
            </p>
          </div>
        </div>
      </template>
      <template #footer>
        <template v-if="!shareResult">
          <el-button @click="showShareDialog = false">取消</el-button>
          <el-button type="primary" :loading="sharing" @click="handleCreateShare">
            创建分享
          </el-button>
        </template>
        <template v-else>
          <el-button type="primary" @click="showShareDialog = false">完成</el-button>
        </template>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useFileStore } from '../stores/file'
import { useUserStore } from '../stores/user'
import { uploadFile, createFolder, deleteFile, downloadFile } from '../api/file'
import { createShare } from '../api/share'

const fileStore = useFileStore()
const userStore = useUserStore()

const showUploadDialog = ref(false)
const showFolderDialog = ref(false)
const showShareDialog = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const uploadFileList = ref([])
const selectedFile = ref(null)
const newFolderName = ref('')
const creatingFolder = ref(false)
const sharing = ref(false)
const shareTarget = ref(null)
const shareResult = ref(null)
const shareForm = reactive({
  expiretime: '7',
  key: '',
})

const shareLink = ref('')

function formatTime(str) {
  if (!str) return '-'
  const d = new Date(str)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function handleRowDblClick(row) {
  if (row.is_folder) {
    fileStore.enterFolder(row)
  }
}

function handleFileChange(file) {
  selectedFile.value = file.raw
  uploadFileList.value = [file]
}

async function handleUpload() {
  if (!selectedFile.value) {
    ElMessage.warning('请选择文件')
    return
  }
  uploading.value = true
  uploadProgress.value = 0
  try {
    await uploadFile(selectedFile.value, fileStore.currentParentId, (e) => {
      if (e.total) {
        uploadProgress.value = Math.round((e.loaded / e.total) * 100)
      }
    })
    ElMessage.success('上传成功')
    showUploadDialog.value = false
    selectedFile.value = null
    uploadFileList.value = []
    fileStore.refresh()
    userStore.fetchUserInfo()
  } catch {
    /* handled */
  } finally {
    uploading.value = false
  }
}

async function handleCreateFolder() {
  const name = newFolderName.value.trim()
  if (!name) {
    ElMessage.warning('请输入文件夹名称')
    return
  }
  creatingFolder.value = true
  try {
    await createFolder(name, fileStore.currentParentId)
    ElMessage.success('文件夹创建成功')
    showFolderDialog.value = false
    newFolderName.value = ''
    fileStore.refresh()
  } catch {
    /* handled */
  } finally {
    creatingFolder.value = false
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除「${row.file_name}」吗？删除后可在回收站恢复。`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await deleteFile(row.id)
    ElMessage.success('已移入回收站')
    fileStore.refresh()
    userStore.fetchUserInfo()
  } catch {
    /* cancelled or error */
  }
}

async function handleDownload(row) {
  try {
    const res = await downloadFile(row.id)
    const url = res.data.url
    if (url) {
      const a = document.createElement('a')
      a.href = url
      a.download = res.data.file_name || row.file_name
      a.target = '_blank'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    }
  } catch {
    /* handled */
  }
}

function handleShare(row) {
  shareTarget.value = row
  shareResult.value = null
  shareForm.expiretime = '7'
  shareForm.key = ''
  shareLink.value = ''
  showShareDialog.value = true
}

async function handleCreateShare() {
  sharing.value = true
  try {
    const payload = {
      user_file_id: shareTarget.value.id,
      expiretime: shareForm.expiretime,
    }
    if (shareForm.key.trim()) {
      payload.key = shareForm.key.trim()
    }
    const res = await createShare(payload)
    shareResult.value = res.data
    shareLink.value = `${window.location.origin}/share/${res.data.shareToken}`
  } catch {
    /* handled */
  } finally {
    sharing.value = false
  }
}

function copyShareLink() {
  let text = shareLink.value
  if (shareForm.key) {
    text += `\n提取码：${shareForm.key}`
  }
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制到剪贴板')
  })
}

onMounted(() => {
  fileStore.fetchFiles(0)
})
</script>

<style scoped>
.home {
  max-width: 1200px;
  margin: 0 auto;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.toolbar-left {
  display: flex;
  gap: 8px;
}

.breadcrumb-bar {
  margin-bottom: 16px;
  padding: 12px 16px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid var(--wd-border);
}

.breadcrumb-link {
  cursor: pointer;
  color: var(--wd-text-secondary);
  transition: color 0.2s;
}

.breadcrumb-link:hover {
  color: var(--wd-primary);
}

.breadcrumb-link.active {
  color: var(--wd-text);
  font-weight: 500;
  cursor: default;
}

.file-table {
  border-radius: 8px;
  overflow: hidden;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.file-name {
  font-size: 14px;
  color: var(--wd-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-meta {
  font-size: 13px;
  color: var(--wd-text-secondary);
}

.share-result {
  padding: 8px 0;
}

.share-link-box {
  margin-top: 16px;
}

.share-key {
  margin-top: 12px;
  font-size: 14px;
  color: var(--wd-text-secondary);
}

.share-key strong {
  color: var(--wd-primary);
  font-size: 16px;
  letter-spacing: 2px;
}
</style>
