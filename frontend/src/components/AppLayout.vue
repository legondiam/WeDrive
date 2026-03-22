<template>
  <div class="flex min-h-screen bg-background">
    <aside class="flex w-56 shrink-0 flex-col border-r border-border bg-card p-3">
      <div class="mb-3 flex items-center gap-2 rounded-md border border-border bg-white px-3 py-3 shadow-sm">
        <BrandCloudIcon class="h-7 w-7 text-slate-500" />
        <span class="font-brand text-[24px] font-bold leading-[1.15] text-zinc-900">WeDrive</span>
      </div>

      <nav class="space-y-1">
        <RouterLink
          v-for="item in menus"
          :key="item.path"
          :to="item.path"
          :class="[
            'flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium leading-[1.6] transition-all duration-200',
            activeMenu === item.path
              ? 'border-neutral-300 bg-neutral-100 text-zinc-900 shadow-sm'
              : 'border-transparent text-zinc-700 hover:border-neutral-200 hover:bg-neutral-100 hover:text-zinc-900',
          ]"
        >
          <component :is="item.icon" class="h-4 w-4" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <Card class="mt-auto p-3">
        <div class="mb-2 flex items-center justify-between text-[12px] leading-[1.6] text-neutral-500">
          <span>空间使用</span>
          <span>{{ userStore.getUsagePercent() }}%</span>
        </div>
        <Progress :value="userStore.getUsagePercent()" />
        <p class="mt-2 text-[12px] leading-[1.6] text-neutral-600">
          {{ userStore.usedSpace || '0B' }} / {{ userStore.totalSpace || '0B' }}
        </p>
      </Card>
    </aside>

    <section class="flex min-h-screen flex-1 flex-col">
      <header class="sticky top-0 z-10 flex h-16 items-center justify-between border-b border-border bg-white/90 px-6 backdrop-blur-sm">
        <div>
          <h2 class="text-[22px] font-bold leading-[1.25] text-zinc-900">{{ pageTitle }}</h2>
          <p class="text-[12px] leading-[1.6] text-zinc-500">{{ pageSubtitle }}</p>
        </div>
        <div class="relative">
          <button
            class="flex items-center gap-2 rounded-md border border-border bg-white px-3 py-1.5 text-[14px] leading-[1.6] shadow-sm transition-all duration-200 hover:border-neutral-300 hover:shadow-md"
            @click="menuOpen = !menuOpen"
          >
            <CircleUserRound class="h-4 w-4" />
            <span>{{ displayName }}</span>
            <ChevronDown class="h-4 w-4" />
          </button>
          <div
            v-if="menuOpen"
            class="absolute right-0 top-11 w-44 rounded-md border border-border bg-white p-1 shadow-md"
          >
            <button
              class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-[14px] leading-[1.6] text-neutral-700 transition-colors hover:bg-neutral-100"
              @click="handleLogout"
            >
              <LogOut class="h-4 w-4" />
              退出登录
            </button>
          </div>
        </div>
      </header>

      <main class="mx-auto w-full max-w-[1320px] flex-1 px-6 py-6">
        <router-view />
      </main>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { FolderOpen, Trash2, Shield, CircleUserRound, ChevronDown, LogOut } from 'lucide-vue-next'
import Card from '@/components/ui/card/Card.vue'
import Progress from '@/components/ui/progress/Progress.vue'
import BrandCloudIcon from '@/components/icons/BrandCloudIcon.vue'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const menuOpen = ref(false)

const menus = [
  { path: '/drive', label: '我的文件', icon: FolderOpen },
  { path: '/drive/recycle', label: '回收站', icon: Trash2 },
  { path: '/drive/admin', label: '管理', icon: Shield },
]

const activeMenu = computed(() => {
  if (route.path === '/drive/recycle') return '/drive/recycle'
  if (route.path === '/drive/admin') return '/drive/admin'
  return '/drive'
})

const pageTitle = computed(() => {
  if (route.path === '/drive/recycle') return '回收站'
  if (route.path === '/drive/admin') return '管理面板'
  return '我的文件'
})

const pageSubtitle = computed(() => {
  if (route.path === '/drive/recycle') return '集中管理已删除文件'
  if (route.path === '/drive/admin') return '更新用户会员状态'
  return '管理你的文件目录与分享内容'
})

const displayName = computed(() => userStore.username || '用户')

function handleLogout() {
  menuOpen.value = false
  userStore.logout()
  router.push('/login')
}

onMounted(() => {
  userStore.fetchUserInfo()
})
</script>
