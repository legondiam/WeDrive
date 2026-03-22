<template>
  <div class="space-y-3">
    <Card class="p-4">
      <h2 class="text-[26px] font-bold leading-[1.2] text-zinc-900">管理面板</h2>
      <p class="mt-1 text-[14px] leading-[1.6] text-zinc-500">管理用户会员状态</p>
    </Card>

    <Card class="p-4">
      <h3 class="text-[22px] font-bold leading-[1.25] text-zinc-900">更新用户会员</h3>

      <div class="mt-4 max-w-xl space-y-4">
        <div class="space-y-1">
          <label class="text-[12px] leading-[1.6] text-neutral-600">目标用户 ID</label>
          <Input v-model="form.target_user_id" type="number" :min="1" placeholder="请输入用户 ID" />
        </div>

        <div class="space-y-1">
          <label class="text-[12px] leading-[1.6] text-neutral-600">会员等级</label>
          <div class="flex gap-2">
            <Button :variant="form.member_level === 0 ? 'default' : 'outline'" size="sm" @click="form.member_level = 0">普通用户</Button>
            <Button :variant="form.member_level === 1 ? 'default' : 'outline'" size="sm" @click="form.member_level = 1">VIP</Button>
          </div>
        </div>

        <div class="space-y-1">
          <label class="text-[12px] leading-[1.6] text-neutral-600">VIP 时长</label>
          <div class="flex gap-2">
            <Button :variant="form.vip_months === 1 ? 'default' : 'outline'" size="sm" :disabled="form.member_level === 0" @click="form.vip_months = 1">1个月</Button>
            <Button :variant="form.vip_months === 3 ? 'default' : 'outline'" size="sm" :disabled="form.member_level === 0" @click="form.vip_months = 3">3个月</Button>
            <Button :variant="form.vip_months === 12 ? 'default' : 'outline'" size="sm" :disabled="form.member_level === 0" @click="form.vip_months = 12">12个月</Button>
          </div>
        </div>

        <Button :disabled="loading" @click="handleSubmit">
          {{ loading ? '更新中...' : '确认更新' }}
        </Button>
      </div>
    </Card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { toast } from 'vue-sonner'
import Card from '@/components/ui/card/Card.vue'
import Input from '@/components/ui/input/Input.vue'
import Button from '@/components/ui/button/Button.vue'
import { updateMember } from '../api/admin'

const loading = ref(false)

const form = reactive({
  target_user_id: undefined,
  member_level: 0,
  vip_months: 1,
})

async function handleSubmit() {
  if (!form.target_user_id || Number(form.target_user_id) < 1) {
    toast.warning('请输入正确的用户 ID')
    return
  }

  loading.value = true
  try {
    await updateMember({
      target_user_id: Number(form.target_user_id),
      member_level: form.member_level,
      vip_months: form.vip_months,
    })
    toast.success('用户会员状态已更新')
  } catch {
    /* handled */
  } finally {
    loading.value = false
  }
}
</script>
