import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUserInfo } from '../api/user'

export const useUserStore = defineStore('user', () => {
  const totalSpace = ref('')
  const usedSpace = ref('')
  const username = ref('')
  const isMember = ref(false)
  const memberStatus = ref('非会员')

  async function fetchUserInfo() {
    try {
      const res = await getUserInfo()
      username.value = res.data.Username || ''
      totalSpace.value = res.data.TotalSpace
      usedSpace.value = res.data.UsedSpace
      isMember.value = Boolean(res.data.IsMember)
      memberStatus.value = res.data.MemberStatus || (isMember.value ? '会员' : '非会员')
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
    isMember.value = false
    memberStatus.value = '非会员'
  }

  return { totalSpace, usedSpace, username, isMember, memberStatus, fetchUserInfo, getUsagePercent, logout }
})
