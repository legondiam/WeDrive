<template>
  <div class="flex min-h-screen items-center justify-center bg-background p-6">
    <Card class="w-full max-w-[400px] p-7">
      <div class="mb-6 text-center">
        <Cloud class="mx-auto h-7 w-7 text-foreground" />
        <h1 class="mt-3 text-[32px] font-semibold leading-[1.4] text-foreground">WeDrive</h1>
        <p class="mt-1 text-[14px] leading-[1.6] text-neutral-500">登录你的云盘账户</p>
      </div>

      <form class="space-y-3" @submit.prevent="handleLogin">
        <div class="space-y-1">
          <label class="text-[12px] leading-[1.6] text-neutral-600">用户名</label>
          <Input v-model="form.username" placeholder="请输入用户名" />
        </div>
        <div class="space-y-1">
          <label class="text-[12px] leading-[1.6] text-neutral-600">密码</label>
          <Input v-model="form.password" type="password" placeholder="请输入密码" />
        </div>
        <Button type="submit" class="w-full" :disabled="loading">
          {{ loading ? '登录中...' : '登录' }}
        </Button>
      </form>

      <p class="mt-4 text-center text-[12px] leading-[1.6] text-neutral-500">
        还没有账户？
        <router-link class="font-medium text-foreground hover:text-neutral-700" to="/register">立即注册</router-link>
      </p>
    </Card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { Cloud } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import Card from '@/components/ui/card/Card.vue'
import Input from '@/components/ui/input/Input.vue'
import Button from '@/components/ui/button/Button.vue'
import { login } from '../api/user'

const router = useRouter()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

async function handleLogin() {
  if (!form.username.trim()) {
    toast.warning('请输入用户名')
    return
  }
  if (!form.password.trim()) {
    toast.warning('请输入密码')
    return
  }
  if (form.password.length < 6) {
    toast.warning('密码至少6个字符')
    return
  }

  loading.value = true
  try {
    const res = await login(form)
    localStorage.setItem('accessToken', res.data.accessToken)
    toast.success('登录成功')
    router.push('/')
  } catch {
    /* error handled by interceptor */
  } finally {
    loading.value = false
  }
}
</script>
