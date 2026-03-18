<template>
  <el-container class="app-layout">
    <el-aside class="sidebar" width="220px">
      <div class="logo">
        <el-icon :size="28" color="#409eff"><Cloudy /></el-icon>
        <span class="logo-text">WeDrive</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/">
          <el-icon><FolderOpened /></el-icon>
          <span>我的文件</span>
        </el-menu-item>
        <el-menu-item index="/recycle">
          <el-icon><Delete /></el-icon>
          <span>回收站</span>
        </el-menu-item>
        <el-menu-item index="/admin">
          <el-icon><Setting /></el-icon>
          <span>管理</span>
        </el-menu-item>
      </el-menu>
      <div class="storage-info">
        <div class="storage-label">
          <span>存储空间</span>
          <span class="storage-detail">{{ userStore.usedSpace || '0B' }} / {{ userStore.totalSpace || '0B' }}</span>
        </div>
        <el-progress
          :percentage="userStore.getUsagePercent()"
          :stroke-width="6"
          :show-text="false"
          :color="progressColor"
        />
      </div>
    </el-aside>
    <el-container class="main-container">
      <el-header class="topbar">
        <div class="topbar-left"></div>
        <div class="topbar-right">
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-dropdown">
              <el-avatar :size="32" class="user-avatar">
                <el-icon :size="18"><User /></el-icon>
              </el-avatar>
              <el-icon class="arrow"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  <span style="color: #303133; font-weight: 500;">{{ displayName }}</span>
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const activeMenu = computed(() => {
  if (route.path === '/recycle') return '/recycle'
  if (route.path === '/admin') return '/admin'
  return '/'
})

const displayName = computed(() => {
  return userStore.username || '用户'
})

const progressColor = computed(() => {
  const p = userStore.getUsagePercent()
  if (p >= 90) return '#f56c6c'
  if (p >= 70) return '#e6a23c'
  return '#409eff'
})

function handleCommand(cmd) {
  if (cmd === 'logout') {
    userStore.logout()
    router.push('/login')
  }
}

onMounted(() => {
  userStore.fetchUserInfo()
})
</script>

<style scoped>
.app-layout {
  height: 100vh;
}

.sidebar {
  background: #fff;
  border-right: 1px solid var(--wd-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 24px;
  border-bottom: 1px solid var(--wd-border);
}

.logo-text {
  font-size: 20px;
  font-weight: 700;
  color: #303133;
  letter-spacing: -0.5px;
}

.sidebar-menu {
  border-right: none;
  flex: 1;
  padding-top: 8px;
}

.sidebar-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
  margin: 2px 8px;
  border-radius: 8px;
}

.sidebar-menu .el-menu-item.is-active {
  background: var(--wd-primary-light);
}

.storage-info {
  padding: 16px 20px;
  border-top: 1px solid var(--wd-border);
}

.storage-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--wd-text-secondary);
}

.storage-detail {
  font-weight: 500;
  color: var(--wd-text);
}

.main-container {
  background: var(--wd-bg);
}

.topbar {
  background: #fff;
  border-bottom: 1px solid var(--wd-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 56px;
}

.topbar-right {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.user-avatar {
  background: var(--wd-primary-light);
  color: var(--wd-primary);
}

.arrow {
  color: var(--wd-text-secondary);
  font-size: 12px;
}

.content {
  padding: 24px;
  overflow-y: auto;
}
</style>
