<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <el-icon :size="36" color="#409eff"><Cloudy /></el-icon>
        <h1>WeDrive</h1>
        <p>登录你的云盘账户</p>
      </div>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        size="large"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            :prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            class="auth-btn"
            @click="handleLogin"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
      <div class="auth-footer">
        还没有账户？<router-link to="/register">立即注册</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '../api/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6个字符', trigger: 'blur' },
  ],
}

async function handleLogin() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await login(form)
    localStorage.setItem('accessToken', res.data.accessToken)
    ElMessage.success('登录成功')
    router.push('/')
  } catch {
    /* error handled by interceptor */
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.auth-card {
  background: #fff;
  border-radius: 16px;
  padding: 48px 40px 36px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.auth-header {
  text-align: center;
  margin-bottom: 32px;
}

.auth-header h1 {
  margin: 12px 0 4px;
  font-size: 28px;
  font-weight: 700;
  color: #303133;
}

.auth-header p {
  color: #909399;
  font-size: 14px;
}

.auth-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  border-radius: 8px;
}

.auth-footer {
  text-align: center;
  margin-top: 16px;
  font-size: 14px;
  color: #909399;
}

.auth-footer a {
  color: #409eff;
  font-weight: 500;
}
</style>
