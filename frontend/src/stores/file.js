import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getFileList } from '../api/file'

export const useFileStore = defineStore('file', () => {
  const files = ref([])
  const loading = ref(false)
  const currentParentId = ref(0)
  const pathStack = ref([{ id: 0, name: '全部文件' }])

  const breadcrumbs = computed(() => pathStack.value)

  async function fetchFiles(parentId) {
    loading.value = true
    try {
      const res = await getFileList(parentId)
      files.value = res.data || []
      currentParentId.value = parentId
    } finally {
      loading.value = false
    }
  }

  function enterFolder(folder) {
    pathStack.value.push({ id: folder.id, name: folder.file_name })
    fetchFiles(folder.id)
  }

  function navigateTo(index) {
    const target = pathStack.value[index]
    pathStack.value = pathStack.value.slice(0, index + 1)
    fetchFiles(target.id)
  }

  function refresh() {
    fetchFiles(currentParentId.value)
  }

  function reset() {
    files.value = []
    currentParentId.value = 0
    pathStack.value = [{ id: 0, name: '全部文件' }]
  }

  return {
    files, loading, currentParentId, breadcrumbs,
    fetchFiles, enterFolder, navigateTo, refresh, reset,
  }
})
