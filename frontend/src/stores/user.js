import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUserInfo } from '../api/user'

export const useUserStore = defineStore('user', () => {
  const totalSpace = ref('')
  const usedSpace = ref('')
  const username = ref('')

  async function fetchUserInfo() {
    try {
      const res = await getUserInfo()
      totalSpace.value = res.data.TotalSpace
      usedSpace.value = res.data.UsedSpace
    } catch {
      /* ignore */
    }
  }

  function parseSize(sizeStr) {
    if (!sizeStr) return 0
    const units = { B: 1, KB: 1024, MB: 1024 ** 2, GB: 1024 ** 3, TB: 1024 ** 4 }
    const match = sizeStr.match(/([\d.]+)\s*(B|KB|MB|GB|TB)/i)
    if (!match) return 0
    return parseFloat(match[1]) * (units[match[2].toUpperCase()] || 1)
  }

  function getUsagePercent() {
    const used = parseSize(usedSpace.value)
    const total = parseSize(totalSpace.value)
    if (total === 0) return 0
    return Math.round((used / total) * 100)
  }

  function logout() {
    localStorage.removeItem('accessToken')
    totalSpace.value = ''
    usedSpace.value = ''
    username.value = ''
  }

  return { totalSpace, usedSpace, username, fetchUserInfo, getUsagePercent, logout }
})
