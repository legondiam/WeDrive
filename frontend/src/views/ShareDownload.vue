<template>
  <div class="flex min-h-screen items-center justify-center bg-background p-6">
    <Card class="w-full max-w-[440px] p-8 text-center">
      <Cloud class="mx-auto h-10 w-10 text-foreground" />
      <h1 class="mt-3 text-[24px] font-semibold leading-[1.4] text-foreground">WeDrive 文件分享</h1>

      <form v-if="!downloadReady" class="mt-6 space-y-3" @submit.prevent="handleGetFile">
        <Input
          v-model="key"
          placeholder="请输入提取码（无密码则留空）"
          maxlength="4"
        />
        <p v-if="errorMessage" class="text-left text-[12px] leading-[1.6] text-red-600">
          {{ errorMessage }}
        </p>
        <Button class="w-full" type="submit" :disabled="loading">
          {{ loading ? '获取中...' : '获取文件' }}
        </Button>
      </form>

      <div v-else class="mt-6 space-y-3">
        <CheckCircle2 class="mx-auto h-10 w-10 text-foreground" />
        <h2 class="break-all text-[20px] font-semibold leading-[1.4] text-foreground">{{ fileName }}</h2>
        <p class="text-[14px] leading-[1.6] text-neutral-500">文件已就绪，点击下方按钮下载</p>
        <Button class="w-full" @click="doDownload">
          <Download class="h-4 w-4" />
          下载文件
        </Button>
      </div>

      <p class="mt-6 text-[12px] leading-[1.6] text-neutral-500">
        <router-link class="hover:text-neutral-700" to="/login">登录 WeDrive</router-link>
      </p>
    </Card>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import { Cloud, Download, CheckCircle2 } from 'lucide-vue-next'
import Card from '@/components/ui/card/Card.vue'
import Input from '@/components/ui/input/Input.vue'
import Button from '@/components/ui/button/Button.vue'
import { downloadShare } from '../api/share'

const route = useRoute()
const shareToken = computed(() => {
  const value = route.params.token
  if (Array.isArray(value)) return value[0] || ''
  return value ? String(value) : ''
})
const key = ref('')
const loading = ref(false)
const downloadReady = ref(false)
const downloadUrl = ref('')
const fileName = ref('')
const errorMessage = ref('')

async function handleGetFile() {
  if (!shareToken.value) {
    errorMessage.value = '分享链接无效'
    toast.error('分享链接无效')
    return
  }

  loading.value = true
  errorMessage.value = ''
  try {
    const payload = { share_token: shareToken.value }
    if (key.value.trim()) {
      payload.key = key.value.trim()
    }
    const res = await downloadShare(payload)
    const url = res?.data?.URL || res?.data?.url || ''
    const name = res?.data?.FileName || res?.data?.fileName || ''
    if (!url) {
      errorMessage.value = '未获取到下载链接'
      toast.error('未获取到下载链接')
      return
    }
    downloadUrl.value = url
    fileName.value = name
    downloadReady.value = true
  } catch (err) {
    errorMessage.value = err?.message || '获取文件失败'
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
