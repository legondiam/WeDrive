<template>
  <div class="space-y-3">
    <Card class="p-4">
      <h2 class="text-[24px] font-semibold leading-[1.4] text-foreground">回收站</h2>
      <p class="mt-1 text-[14px] leading-[1.6] text-neutral-500">已删除的文件将在此处保留，可恢复或永久删除</p>
    </Card>

    <Card class="overflow-hidden">
      <div v-if="loading" class="p-6 text-center text-[14px] leading-[1.6] text-neutral-500">加载中...</div>
      <div v-else-if="!files.length" class="p-6 text-center text-[14px] leading-[1.6] text-neutral-500">回收站为空</div>
      <table v-else class="w-full border-collapse">
        <thead class="bg-neutral-50">
          <tr>
            <th class="px-4 py-3 text-left text-[12px] font-medium leading-[1.6] text-neutral-600">名称</th>
            <th class="px-4 py-3 text-center text-[12px] font-medium leading-[1.6] text-neutral-600">大小</th>
            <th class="px-4 py-3 text-center text-[12px] font-medium leading-[1.6] text-neutral-600">删除日期</th>
            <th class="px-4 py-3 text-right text-[12px] font-medium leading-[1.6] text-neutral-600">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in files" :key="row.id" class="border-t border-border hover:bg-neutral-50">
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <Folder v-if="row.is_folder" class="h-4 w-4 text-neutral-600" />
                <FileText v-else class="h-4 w-4 text-neutral-500" />
                <span class="max-w-[360px] truncate text-[14px] font-medium leading-[1.6] text-neutral-900">{{ row.file_name }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-center text-[12px] leading-[1.6] text-neutral-500">{{ row.is_folder ? '-' : row.file_size }}</td>
            <td class="px-4 py-3 text-center text-[12px] leading-[1.6] text-neutral-500">{{ formatTime(row.deleted_at) }}</td>
            <td class="px-4 py-3">
              <div class="flex items-center justify-end gap-1">
                <Button variant="ghost" size="sm" @click="handleRestore(row)">恢复</Button>
                <Button variant="ghost" size="sm" class="text-neutral-700" @click="handlePermanentDelete(row)">彻底删除</Button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { Folder, FileText } from 'lucide-vue-next'
import Card from '@/components/ui/card/Card.vue'
import Button from '@/components/ui/button/Button.vue'
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
  await restoreFile(row.id)
  toast.success(`「${row.file_name}」已恢复`)
  fetchRecycle()
  userStore.fetchUserInfo()
}

async function handlePermanentDelete(row) {
  const ok = window.confirm(`确定要彻底删除「${row.file_name}」吗？此操作不可恢复！`)
  if (!ok) return
  await permanentDeleteFile(row.id)
  toast.success('已彻底删除')
  fetchRecycle()
  userStore.fetchUserInfo()
}

onMounted(fetchRecycle)
</script>
