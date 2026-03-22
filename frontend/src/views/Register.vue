<template>
  <div class="flex min-h-screen items-center justify-center bg-background p-6">
    <div class="relative w-full max-w-[400px]">
      <Card class="auth-shine-card w-full overflow-hidden p-7">
        <div class="mb-6 text-center">
          <BrandCloudIcon class="mx-auto h-9 w-9 text-slate-500" />
          <h1 class="font-brand mt-3 text-[24px] font-bold leading-[1.15] text-zinc-900">WeDrive</h1>
          <p class="mt-1 text-[14px] leading-[1.6] text-zinc-500">创建一个新账户</p>
        </div>

        <form class="space-y-3" @submit.prevent="handleRegister">
          <div class="space-y-1">
            <label class="text-[12px] leading-[1.6] text-neutral-600">用户名</label>
            <Input v-model="form.username" placeholder="请输入用户名" />
          </div>
          <div class="space-y-1">
            <label class="text-[12px] leading-[1.6] text-neutral-600">密码</label>
            <Input v-model="form.password" type="password" placeholder="请输入密码" />
          </div>
          <div class="space-y-1">
            <label class="text-[12px] leading-[1.6] text-neutral-600">确认密码</label>
            <Input v-model="form.confirmPassword" type="password" placeholder="请再次输入密码" />
          </div>
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? '注册中...' : '注册' }}
          </Button>
        </form>

        <p class="mt-4 text-center text-[12px] leading-[1.6] text-neutral-500">
          已有账户？
          <router-link class="font-medium text-foreground hover:text-neutral-700" to="/login">立即登录</router-link>
        </p>
      </Card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import Card from '@/components/ui/card/Card.vue'
import Input from '@/components/ui/input/Input.vue'
import Button from '@/components/ui/button/Button.vue'
import BrandCloudIcon from '@/components/icons/BrandCloudIcon.vue'
import { register } from '../api/user'

const router = useRouter()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
})

async function handleRegister() {
  if (!form.username.trim()) {
    toast.warning('请输入用户名')
    return
  }
  if (form.username.length < 3) {
    toast.warning('用户名至少3个字符')
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
  if (form.confirmPassword !== form.password) {
    toast.warning('两次密码不一致')
    return
  }

  loading.value = true
  try {
    await register({
      username: form.username,
      password: form.password,
    })
    toast.success('注册成功，请登录')
    router.push('/login')
  } catch {
    /* error handled by interceptor */
  } finally {
    loading.value = false
  }
}
</script>
