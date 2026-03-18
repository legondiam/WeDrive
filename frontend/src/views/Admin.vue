<template>
  <div class="admin">
    <div class="page-header">
      <h2>管理面板</h2>
      <span class="page-desc">管理用户会员状态</span>
    </div>

    <el-card class="admin-card" shadow="never">
      <template #header>
        <span>更新用户会员</span>
      </template>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        style="max-width: 500px"
      >
        <el-form-item label="目标用户 ID" prop="target_user_id">
          <el-input-number
            v-model="form.target_user_id"
            :min="1"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item label="会员等级" prop="member_level">
          <el-radio-group v-model="form.member_level">
            <el-radio-button :value="0">普通用户</el-radio-button>
            <el-radio-button :value="1">VIP</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="VIP 时长" prop="vip_months">
          <el-radio-group v-model="form.vip_months" :disabled="form.member_level === 0">
            <el-radio-button :value="1">1个月</el-radio-button>
            <el-radio-button :value="3">3个月</el-radio-button>
            <el-radio-button :value="12">12个月</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSubmit">
            确认更新
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { updateMember } from '../api/admin'

const formRef = ref()
const loading = ref(false)

const form = reactive({
  target_user_id: undefined,
  member_level: 0,
  vip_months: 1,
})

const rules = {
  target_user_id: [{ required: true, message: '请输入用户 ID', trigger: 'blur' }],
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await updateMember(form)
    ElMessage.success('用户会员状态已更新')
  } catch {
    /* handled */
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.admin {
  max-width: 800px;
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

.admin-card {
  border-radius: 8px;
}
</style>
