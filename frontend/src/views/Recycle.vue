<template>
  <div class="recycle">
    <div class="page-header">
      <h2>回收站</h2>
      <span class="page-desc">已删除的文件将在此处保留，可恢复或永久删除</span>
    </div>

    <el-table
      v-loading="loading"
      :data="files"
      class="file-table"
      empty-text="回收站为空"
    >
      <el-table-column label="名称" min-width="300">
        <template #default="{ row }">
          <div class="file-name-cell">
            <el-icon :size="22" :color="row.is_folder ? '#e6a23c' : '#909399'">
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
      <el-table-column label="删除日期" width="180" align="center">
        <template #default="{ row }">
          <span class="file-meta">{{ formatTime(row.deleted_at) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" align="center" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleRestore(row)">
            <el-icon><RefreshLeft /></el-icon>
            恢复
          </el-button>
          <el-button link type="danger" size="small" @click="handlePermanentDelete(row)">
            <el-icon><Delete /></el-icon>
            彻底删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRecycleList, restoreFile, permanentDeleteFile } from '../api/file'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()
const loading = ref(false)
const files = ref([])

function formatTime(str) {
  if (!str) return '-'
  const d = new Date(str)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

async function fetchRecycle() {
  loading.value = true
  try {
    const res = await getRecycleList()
    files.value = res.data || []
  } finally {
    loading.value = false
  }
}

async function handleRestore(row) {
  try {
    await restoreFile(row.id)
    ElMessage.success(`「${row.file_name}」已恢复`)
    fetchRecycle()
    userStore.fetchUserInfo()
  } catch {
    /* handled */
  }
}

async function handlePermanentDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定要彻底删除「${row.file_name}」吗？此操作不可恢复！`,
      '彻底删除',
      { confirmButtonText: '彻底删除', cancelButtonText: '取消', type: 'error' }
    )
    await permanentDeleteFile(row.id)
    ElMessage.success('已彻底删除')
    fetchRecycle()
    userStore.fetchUserInfo()
  } catch {
    /* cancelled or error */
  }
}

onMounted(fetchRecycle)
</script>

<style scoped>
.recycle {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 4px;
}

.page-desc {
  font-size: 13px;
  color: var(--wd-text-secondary);
}

.file-table {
  border-radius: 8px;
  overflow: hidden;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.file-name {
  font-size: 14px;
  color: var(--wd-text);
}

.file-meta {
  font-size: 13px;
  color: var(--wd-text-secondary);
}
</style>
