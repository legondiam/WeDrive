<template>
  <div class="flex min-h-screen items-center justify-center bg-background p-6">
    <Card class="w-full max-w-[440px] p-8 text-center">
      <Cloud class="mx-auto h-10 w-10 text-foreground" />
      <h1 class="mt-3 text-[24px] font-semibold leading-[1.4] text-foreground">WeDrive 文件分享</h1>

      <div v-if="!downloadReady" class="mt-6 space-y-3">
        <Input
          v-model="key"
          placeholder="请输入提取码（无密码则留空）"
          maxlength="4"
          @keyup.enter="handleGetFile"
        />
        <Button class="w-full" :disabled="loading" @click="handleGetFile">
          {{ loading ? '获取中...' : '获取文件' }}
        </Button>
      </div>

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
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { Cloud, Download, CheckCircle2 } from 'lucide-vue-next'
import Card from '@/components/ui/card/Card.vue'
import Input from '@/components/ui/input/Input.vue'
import Button from '@/components/ui/button/Button.vue'
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
