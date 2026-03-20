<template>
  <div class="space-y-3">
    <Card class="p-3">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <Dialog v-model:open="showUploadDialog">
            <DialogTrigger as-child>
              <Button>
                <Upload class="h-4 w-4" />
                上传文件
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>上传文件</DialogTitle>
                <DialogDescription>每次仅支持上传一个文件</DialogDescription>
              </DialogHeader>
              <Form class="mt-4" @submit.prevent="handleUpload">
                <FormItem>
                  <FormLabel>选择文件</FormLabel>
                  <FormControl>
                    <Input type="file" @change="handleFileChange" />
                  </FormControl>
                  <FormDescription>{{ selectedFileName || '未选择文件' }}</FormDescription>
                </FormItem>
                <FormItem v-if="uploading">
                  <FormLabel>上传进度</FormLabel>
                  <Progress :value="uploadProgress" />
                  <FormDescription>{{ uploadProgress }}%</FormDescription>
                </FormItem>
              </Form>
              <DialogFooter>
                <DialogClose as-child>
                  <Button variant="outline" @click="closeUpload">取消</Button>
                </DialogClose>
                <Button :disabled="uploading" @click="handleUpload">
                  {{ uploading ? '上传中...' : '开始上传' }}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          <Dialog v-model:open="showFolderDialog">
            <DialogTrigger as-child>
              <Button variant="outline">
                <FolderPlus class="h-4 w-4" />
                新建文件夹
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>新建文件夹</DialogTitle>
              </DialogHeader>
              <Form class="mt-4" @submit.prevent="handleCreateFolder">
                <FormItem>
                  <FormLabel for="new-folder-input">文件夹名称</FormLabel>
                  <FormControl>
                    <Input
                      id="new-folder-input"
                      v-model="newFolderName"
                      placeholder="请输入文件夹名称"
                      @keyup.enter="handleCreateFolder"
                    />
                  </FormControl>
                </FormItem>
              </Form>
              <DialogFooter>
                <DialogClose as-child>
                  <Button variant="outline">取消</Button>
                </DialogClose>
                <Button :disabled="creatingFolder" @click="handleCreateFolder">创建</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
        <Button size="icon" variant="outline" @click="fileStore.refresh()">
          <RefreshCw class="h-4 w-4" />
        </Button>
      </div>
    </Card>

    <Card class="p-3">
      <div class="flex flex-wrap items-center gap-2 text-[12px] leading-[1.6] text-neutral-500">
        <button
          v-for="(item, index) in fileStore.breadcrumbs"
          :key="item.id"
          class="rounded-md border px-2 py-0.5 transition-colors"
          :class="index === fileStore.breadcrumbs.length - 1
            ? 'border-neutral-300 bg-neutral-100 text-neutral-900'
            : 'border-border bg-white hover:bg-neutral-50'"
          @click="fileStore.navigateTo(index)"
        >
          {{ item.name }}
        </button>
      </div>
    </Card>

    <Card class="overflow-hidden">
      <div v-if="fileStore.loading" class="p-6 text-center text-[14px] leading-[1.6] text-neutral-500">
        加载中...
      </div>
      <div v-else-if="!fileStore.files.length" class="p-6 text-center text-[14px] leading-[1.6] text-neutral-500">
        暂无文件
      </div>
      <Table v-else>
        <TableHeader class="bg-neutral-50">
          <TableRow class="hover:bg-transparent">
            <TableHead>名称</TableHead>
            <TableHead class="text-center">大小</TableHead>
            <TableHead class="text-center">修改时间</TableHead>
            <TableHead class="text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="row in fileStore.files"
            :key="row.id"
            class="cursor-default"
            @dblclick="handleRowDblClick(row)"
          >
            <TableCell>
              <div class="flex items-center gap-2">
                <Folder v-if="row.is_folder" class="h-4 w-4 text-neutral-600" />
                <FileText v-else class="h-4 w-4 text-neutral-500" />
                <span class="max-w-[380px] truncate text-[14px] font-medium leading-[1.6] text-neutral-900">{{ row.file_name }}</span>
              </div>
            </TableCell>
            <TableCell class="text-center text-[12px] leading-[1.6] text-neutral-500">{{ row.is_folder ? '-' : row.file_size }}</TableCell>
            <TableCell class="text-center text-[12px] leading-[1.6] text-neutral-500">{{ formatTime(row.updated_at) }}</TableCell>
            <TableCell>
              <div class="flex items-center justify-end">
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal class="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>文件操作</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem v-if="!row.is_folder" @select.prevent="handleDownload(row)">下载</DropdownMenuItem>
                    <DropdownMenuItem v-if="!row.is_folder" @select.prevent="handleShare(row)">分享</DropdownMenuItem>
                    <DropdownMenuItem class="text-red-600 focus:text-red-600" @select.prevent="handleDelete(row)">
                      删除
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>

    <Dialog v-model:open="showShareDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>创建分享链接</DialogTitle>
        </DialogHeader>

        <Form v-if="!shareResult" class="mt-4">
          <div class="rounded-md border border-border bg-neutral-50 px-3 py-2 text-[14px] leading-[1.6] text-neutral-700">
            文件：{{ shareTarget?.file_name }}
          </div>
          <FormItem>
            <FormLabel>有效期</FormLabel>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="outline" class="w-full justify-between">
                  {{ currentExpireLabel }}
                  <ChevronDown class="h-4 w-4 text-neutral-500" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent class="w-[var(--radix-dropdown-menu-trigger-width)]">
                <DropdownMenuRadioGroup :model-value="shareForm.expiretime">
                  <DropdownMenuRadioItem
                    v-for="opt in expireOptions"
                    :key="opt.value"
                    :value="opt.value"
                    @select.prevent="shareForm.expiretime = opt.value"
                  >
                    {{ opt.label }}
                  </DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </FormItem>
          <FormItem>
            <FormLabel>提取码（可选）</FormLabel>
            <FormControl>
              <Input v-model="shareForm.key" maxlength="4" placeholder="留空则无需提取码" />
            </FormControl>
            <FormDescription>最多 4 位，留空则无需提取码。</FormDescription>
          </FormItem>
        </Form>

        <div v-else class="mt-4 space-y-2">
          <p class="rounded-md border border-border bg-neutral-50 px-3 py-2 text-[14px] leading-[1.6] text-neutral-700">
            分享链接已创建
          </p>
          <Input :model-value="shareLink" readonly />
          <p v-if="shareForm.key" class="text-[12px] leading-[1.6] text-neutral-500">提取码：{{ shareForm.key }}</p>
        </div>

        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline">关闭</Button>
          </DialogClose>
          <Button
            v-if="!shareResult"
            :disabled="sharing"
            @click="handleCreateShare"
          >
            创建分享
          </Button>
          <Button v-else @click="copyShareLink">复制链接</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { toast } from 'vue-sonner'
import { Upload, FolderPlus, RefreshCw, Folder, FileText, MoreHorizontal, ChevronDown } from 'lucide-vue-next'
import Card from '@/components/ui/card/Card.vue'
import Button from '@/components/ui/button/Button.vue'
import Input from '@/components/ui/input/Input.vue'
import Dialog from '@/components/ui/dialog/Dialog.vue'
import DialogTrigger from '@/components/ui/dialog/DialogTrigger.vue'
import DialogContent from '@/components/ui/dialog/DialogContent.vue'
import DialogHeader from '@/components/ui/dialog/DialogHeader.vue'
import DialogTitle from '@/components/ui/dialog/DialogTitle.vue'
import DialogDescription from '@/components/ui/dialog/DialogDescription.vue'
import DialogFooter from '@/components/ui/dialog/DialogFooter.vue'
import DialogClose from '@/components/ui/dialog/DialogClose.vue'
import Progress from '@/components/ui/progress/Progress.vue'
import Form from '@/components/ui/form/Form.vue'
import FormItem from '@/components/ui/form/FormItem.vue'
import FormLabel from '@/components/ui/form/FormLabel.vue'
import FormControl from '@/components/ui/form/FormControl.vue'
import FormDescription from '@/components/ui/form/FormDescription.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import DropdownMenu from '@/components/ui/dropdown-menu/DropdownMenu.vue'
import DropdownMenuTrigger from '@/components/ui/dropdown-menu/DropdownMenuTrigger.vue'
import DropdownMenuContent from '@/components/ui/dropdown-menu/DropdownMenuContent.vue'
import DropdownMenuItem from '@/components/ui/dropdown-menu/DropdownMenuItem.vue'
import DropdownMenuLabel from '@/components/ui/dropdown-menu/DropdownMenuLabel.vue'
import DropdownMenuSeparator from '@/components/ui/dropdown-menu/DropdownMenuSeparator.vue'
import DropdownMenuRadioGroup from '@/components/ui/dropdown-menu/DropdownMenuRadioGroup.vue'
import DropdownMenuRadioItem from '@/components/ui/dropdown-menu/DropdownMenuRadioItem.vue'
import { useFileStore } from '../stores/file'
import { useUserStore } from '../stores/user'
import { uploadFile, createFolder, deleteFile, downloadFile } from '../api/file'
import { createShare } from '../api/share'

const fileStore = useFileStore()
const userStore = useUserStore()

const showUploadDialog = ref(false)
const showFolderDialog = ref(false)
const showShareDialog = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const selectedFile = ref(null)
const selectedFileName = ref('')
const newFolderName = ref('')
const creatingFolder = ref(false)
const sharing = ref(false)
const shareTarget = ref(null)
const shareResult = ref(null)
const shareForm = reactive({
  expiretime: '7',
  key: '',
})
const shareLink = ref('')

const expireOptions = [
  { value: '1', label: '1天' },
  { value: '7', label: '7天' },
  { value: '30', label: '30天' },
  { value: 'permanent', label: '永久' },
]
const currentExpireLabel = computed(() => expireOptions.find((opt) => opt.value === shareForm.expiretime)?.label || '请选择')

function formatTime(str) {
  if (!str) return '-'
  const d = new Date(str)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function handleRowDblClick(row) {
  if (row.is_folder) fileStore.enterFolder(row)
}

function handleFileChange(event) {
  const file = event.target.files?.[0]
  selectedFile.value = file || null
  selectedFileName.value = file?.name || ''
}

function closeUpload() {
  showUploadDialog.value = false
  selectedFile.value = null
  selectedFileName.value = ''
  uploadProgress.value = 0
}

async function handleUpload() {
  if (!selectedFile.value) {
    toast.warning('请选择文件')
    return
  }
  uploading.value = true
  uploadProgress.value = 0
  try {
    await uploadFile(selectedFile.value, fileStore.currentParentId, (e) => {
      if (e.total) uploadProgress.value = Math.round((e.loaded / e.total) * 100)
    })
    toast.success('上传成功')
    closeUpload()
    fileStore.refresh()
    userStore.fetchUserInfo()
  } finally {
    uploading.value = false
  }
}

async function handleCreateFolder() {
  const name = newFolderName.value.trim()
  if (!name) {
    toast.warning('请输入文件夹名称')
    return
  }
  creatingFolder.value = true
  try {
    await createFolder(name, fileStore.currentParentId)
    toast.success('文件夹创建成功')
    showFolderDialog.value = false
    newFolderName.value = ''
    fileStore.refresh()
  } finally {
    creatingFolder.value = false
  }
}

async function handleDelete(row) {
  const ok = window.confirm(`确定要删除「${row.file_name}」吗？删除后可在回收站恢复。`)
  if (!ok) return
  await deleteFile(row.id)
  toast.success('已移入回收站')
  fileStore.refresh()
  userStore.fetchUserInfo()
}

async function handleDownload(row) {
  const res = await downloadFile(row.id)
  const url = res.data.url
  if (!url) return
  const a = document.createElement('a')
  a.href = url
  a.download = res.data.file_name || row.file_name
  a.target = '_blank'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

function handleShare(row) {
  shareTarget.value = row
  shareResult.value = null
  shareForm.expiretime = '7'
  shareForm.key = ''
  shareLink.value = ''
  showShareDialog.value = true
}

async function handleCreateShare() {
  if (!shareTarget.value) return
  sharing.value = true
  try {
    const payload = {
      user_file_id: shareTarget.value.id,
      expiretime: shareForm.expiretime,
    }
    if (shareForm.key.trim()) payload.key = shareForm.key.trim()
    const res = await createShare(payload)
    shareResult.value = res.data
    shareLink.value = `${window.location.origin}/share/${res.data.shareToken}`
  } finally {
    sharing.value = false
  }
}

function copyShareLink() {
  let text = shareLink.value
  if (shareForm.key) text += `\n提取码：${shareForm.key}`
  navigator.clipboard.writeText(text).then(() => {
    toast.success('已复制到剪贴板')
  })
}

onMounted(() => {
  fileStore.fetchFiles(0)
})
</script>
