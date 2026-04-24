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
            <DialogContent v-if="showUploadDialog">
              <DialogHeader>
                <DialogTitle>上传文件</DialogTitle>
              </DialogHeader>
              <Form class="mt-4">
                <FormItem>
                  <FormLabel>选择文件</FormLabel>
                  <div class="mt-2">
                    <FilePond
                      ref="uploadPond"
                      name="file"
                      :credits="false"
                      :allow-multiple="false"
                      :allow-replace="true"
                      :max-files="1"
                      :instant-upload="false"
                      :server="filePondServer"
                      label-idle="拖拽文件到这里或 <span class='filepond--label-action'>点击选择</span>"
                    />
                  </div>
                  <FormDescription>支持拖拽上传，单次仅上传一个文件。</FormDescription>
                </FormItem>
              </Form>
              <DialogFooter>
                <DialogClose as-child>
                  <Button variant="outline" class="transition-all duration-150 ease-out" @click="closeUpload">关闭</Button>
                </DialogClose>
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

          <Dialog v-model:open="showShareDownloadDialog">
            <DialogTrigger as-child>
              <Button variant="outline">
                <ShareExtractIcon class="h-4 w-4" />
                提取分享
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>提取分享文件</DialogTitle>
                <DialogDescription>粘贴分享链接，然后输入提取码获取文件。</DialogDescription>
              </DialogHeader>
              <Form class="mt-4 space-y-4" @submit.prevent="handleFetchSharedFile">
                <FormItem>
                  <FormLabel>分享链接</FormLabel>
                  <FormControl>
                    <Input
                      v-model="shareDownloadForm.link"
                      placeholder="请输入完整分享链接"
                    />
                  </FormControl>
                </FormItem>
                <FormItem>
                  <FormLabel>提取码（可选）</FormLabel>
                  <FormControl>
                    <Input
                      v-model="shareDownloadForm.key"
                      maxlength="4"
                      placeholder="无提取码可留空"
                    />
                  </FormControl>
                </FormItem>
                <p v-if="shareDownloadError" class="text-[12px] leading-[1.6] text-red-600">
                  {{ shareDownloadError }}
                </p>
                <div
                  v-if="sharedDownloadReady"
                  class="rounded-md border border-border bg-neutral-50 px-3 py-3 text-left"
                >
                  <p class="text-[13px] font-medium leading-[1.6] text-neutral-900">{{ sharedDownloadName }}</p>
                  <p class="text-[12px] leading-[1.6] text-neutral-500">文件已就绪，可直接下载。</p>
                </div>
              </Form>
              <DialogFooter>
                <DialogClose as-child>
                  <Button variant="outline" @click="resetShareDownloadDialog">关闭</Button>
                </DialogClose>
                <Button
                  v-if="!sharedDownloadReady"
                  :disabled="shareDownloading"
                  @click="handleFetchSharedFile"
                >
                  {{ shareDownloading ? '获取中...' : '获取文件' }}
                </Button>
                <Button v-else @click="downloadSharedFile">
                  <Download class="h-4 w-4" />
                  下载文件
                </Button>
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
      <div
        v-if="selectedRows.length"
        class="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-stone-50 px-4 py-3"
      >
        <div class="flex items-center gap-3">
          <span class="rounded-full bg-stone-200 px-2.5 py-1 text-[12px] font-medium leading-none text-stone-700">
            已选 {{ selectedRows.length }} 项
          </span>
          <button
            class="text-[13px] leading-[1.6] text-stone-500 transition-colors hover:text-stone-700"
            @click="clearSelection"
          >
            取消选择
          </button>
        </div>
        <Button
          variant="destructive"
          :disabled="deleting"
          @click="handleBatchDelete"
        >
          移入回收站
        </Button>
      </div>
      <div v-if="fileStore.loading" class="p-6 text-center text-[14px] leading-[1.6] text-neutral-500">
        加载中...
      </div>
      <div v-else-if="!fileStore.files.length" class="p-6 text-center text-[14px] leading-[1.6] text-neutral-500">
        暂无文件
      </div>
      <Table v-else>
        <TableHeader class="bg-neutral-50">
          <TableRow class="hover:bg-transparent">
            <TableHead class="w-14 text-center">
              <input
                type="checkbox"
                class="selection-checkbox"
                :checked="allRowsSelected"
                :indeterminate.prop="someRowsSelected"
                @change="toggleSelectAll"
              >
            </TableHead>
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
            class="cursor-default transition-colors"
            :class="isSelected(row.id) ? 'bg-stone-50/80 hover:bg-stone-50' : ''"
            @dblclick="handleRowDblClick(row)"
          >
            <TableCell class="text-center">
              <input
                :checked="isSelected(row.id)"
                type="checkbox"
                class="selection-checkbox"
                @change="toggleRowSelection(row.id, $event.target.checked)"
                @click.stop
              >
            </TableCell>
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

    <Dialog v-model:open="showDeleteDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>删除文件</DialogTitle>
          <DialogDescription>
            {{ deleteDialogDescription }}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline">取消</Button>
          </DialogClose>
          <Button variant="destructive" :disabled="deleting" @click="confirmDelete">
            {{ deleting ? '删除中...' : '确认删除' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup>
import { ref, shallowRef, reactive, onMounted, computed, watch } from 'vue'
import vueFilePond from 'vue-filepond'
import 'filepond/dist/filepond.min.css'
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
import Form from '@/components/ui/form/Form.vue'
import FormItem from '@/components/ui/form/FormItem.vue'
import FormLabel from '@/components/ui/form/FormLabel.vue'
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
import ShareExtractIcon from '@/components/icons/ShareExtractIcon.vue'
import { useFileStore } from '../stores/file'
import { useUserStore } from '../stores/user'
import { uploadFile, quickCheck, prepareInstantUpload, instantUpload, initChunkUpload, signPartUpload, reportUploadedPart, uploadChunkDirect, completeChunkUpload, createFolder, deleteFile, batchDeleteFiles, permanentDeleteFile, downloadFile } from '../api/file'
import { createShare, downloadShare } from '../api/share'
import { calculateFileSHA256, calculateFileSampleSHA256, calculateChunkIdentity, readFileSegmentBase64, CHUNK_IDENTITY_SIZE } from '../lib/sha256'

const CODE_INSTANT_UNAVAILABLE = 3003
const CODE_CHUNK_ALREADY_UPLOADED = 3008
const CODE_CHUNK_HASH_CONFLICT = 3009
const CHUNK_UPLOAD_THRESHOLD = 16 * 1024 * 1024
const HASH_TYPE = 'full_sha256_v1'
const CHUNK_SIZE = CHUNK_IDENTITY_SIZE

const FilePond = vueFilePond()
const fileStore = useFileStore()
const userStore = useUserStore()

const showUploadDialog = ref(false)
const showFolderDialog = ref(false)
const showShareDialog = ref(false)
const showShareDownloadDialog = ref(false)
const showDeleteDialog = ref(false)
const uploadPond = shallowRef(null)
const newFolderName = ref('')
const creatingFolder = ref(false)
const deleting = ref(false)
const sharing = ref(false)
const shareDownloading = ref(false)
const shareTarget = ref(null)
const deleteTarget = ref(null)
const deleteTargets = ref([])
const shareResult = ref(null)
const selectedIds = ref([])
const shareForm = reactive({
  expiretime: '7',
  key: '',
})
const shareDownloadForm = reactive({
  link: '',
  key: '',
})
const shareLink = ref('')
const sharedDownloadUrl = ref('')
const sharedDownloadName = ref('')
const sharedDownloadReady = ref(false)
const shareDownloadError = ref('')

const expireOptions = [
  { value: '1', label: '1天' },
  { value: '7', label: '7天' },
  { value: '30', label: '30天' },
  { value: 'permanent', label: '永久' },
]
const currentExpireLabel = computed(() => expireOptions.find((opt) => opt.value === shareForm.expiretime)?.label || '请选择')
const selectedRows = computed(() => fileStore.files.filter((row) => selectedIds.value.includes(row.id)))
const allRowsSelected = computed(() => fileStore.files.length > 0 && selectedRows.value.length === fileStore.files.length)
const someRowsSelected = computed(() => selectedRows.value.length > 0 && selectedRows.value.length < fileStore.files.length)
const deleteDialogDescription = computed(() => {
  if (deleteTargets.value.length > 1) {
    return `确定要将选中的 ${deleteTargets.value.length} 个项目移入回收站吗？删除后可在回收站恢复。`
  }
  const currentTarget = deleteTarget.value || deleteTargets.value[0]
  return `确定要删除「${currentTarget?.file_name || ''}」吗？删除后可在回收站恢复。`
})

function shouldUseChunkUpload(file) {
  return file.size >= CHUNK_UPLOAD_THRESHOLD
}

const filePondServer = {
  process: (fieldName, file, metadata, load, error, progress, abort) => {
    const controller = new AbortController()
    let finalized = false
    let chunkIdentity = null
    let fullFileHash = null
    const finishUpload = (uploadedId, message, options = {}) => {
      if (finalized) return

      finalized = true
      const serverId = String(uploadedId || Date.now())

      if (!options.instant) {
        progress(true, file.size || 1, file.size || 1)
      }

      // 秒传不会产生真实上传进度，直接结束 FilePond 的处理态。
      load(serverId)
      toast.success(message)
      Promise.allSettled([
        fileStore.refresh(),
        userStore.fetchUserInfo(),
      ])
    }

    calculateFileSampleSHA256(file)
      .then(async (sampleHashes) => {
        if (finalized) return
        const quickCheckRes = await quickCheck(sampleHashes)
        if (finalized) return

        const ensureChunkIdentity = async () => {
          if (!chunkIdentity) {
            chunkIdentity = await calculateChunkIdentity(file, CHUNK_SIZE)
          }
          return chunkIdentity
        }

        const ensureFullFileHash = async () => {
          if (!fullFileHash) {
            fullFileHash = await calculateFileSHA256(file)
          }
          return fullFileHash
        }

        if (quickCheckRes?.data === true) {
          const fileHash = await ensureFullFileHash()
          if (finalized) return

          let instantRes = null
          try {
            const prepareRes = await prepareInstantUpload({
              hash_type: HASH_TYPE,
              file_hash: fileHash,
              file_name: file.name,
              file_size: file.size,
              parent_id: fileStore.currentParentId,
            })
            if (finalized) return

            const prepareData = prepareRes?.data
            const proofs = await Promise.all((prepareData?.challenges || []).map(async (challenge) => ({
              offset: challenge.offset,
              length: challenge.length,
              content_base64: await readFileSegmentBase64(file, challenge.offset, challenge.length),
            })))
            if (finalized) return

            instantRes = await instantUpload({
              hash_type: HASH_TYPE,
              file_hash: fileHash,
              file_name: file.name,
              file_size: file.size,
              parent_id: fileStore.currentParentId,
              prepare_id: prepareData?.prepare_id,
              proofs,
            })
          } catch (err) {
            if (err?.code !== CODE_INSTANT_UNAVAILABLE) throw err
          }
          if (finalized) return

          if (instantRes?.data?.instant) {
            const uploadedId = instantRes?.data?.id
            finishUpload(uploadedId, '秒传成功', { instant: true })
            return
          }
        }

        if (shouldUseChunkUpload(file)) {
          const identity = await ensureChunkIdentity()
          const fileHash = await ensureFullFileHash()
          if (finalized) return

          const initRes = await initChunkUpload({
            hash_type: HASH_TYPE,
            file_hash: fileHash,
            file_name: file.name,
            file_size: file.size,
            parent_id: fileStore.currentParentId,
            chunk_size: identity.chunk_size,
            chunk_count: identity.chunk_count,
            head_hash: sampleHashes.head_hash,
            mid_hash: sampleHashes.mid_hash,
            tail_hash: sampleHashes.tail_hash,
          })
          if (finalized) return

          const uploadId = initRes?.data?.upload_id
          if (!uploadId) {
            throw new Error('分块上传初始化失败')
          }
          const uploadedChunks = new Set(initRes?.data?.uploaded_chunks || [])
          const uploadChunksWithResume = async () => {
            for (const part of identity.parts) {
              const chunkIndex = part.part_number - 1
              const start = chunkIndex * CHUNK_SIZE
              const end = Math.min(file.size, start + CHUNK_SIZE)

              if (uploadedChunks.has(part.part_number)) {
                progress(true, end, file.size || 1)
                continue
              }

              let signRes = null
              try {
                signRes = await signPartUpload({
                  upload_id: uploadId,
                  part_number: part.part_number,
                  chunk_hash: part.chunk_hash,
                })
              } catch (err) {
                if (err?.code === CODE_CHUNK_ALREADY_UPLOADED) {
                  uploadedChunks.add(part.part_number)
                  progress(true, end, file.size || 1)
                  continue
                }
                if (err?.code === CODE_CHUNK_HASH_CONFLICT) {
                  throw new Error('文件内容已变化，请重新开始上传')
                }
                throw err
              }
              if (finalized) return null

              const directRes = await uploadChunkDirect(
                signRes?.data?.upload_url,
                part.chunk,
                signRes?.data?.headers || {},
                (event) => {
                  if (finalized) return
                  const currentLoaded = event.loaded || 0
                  const uploadedBefore = start
                  const total = file.size || 1
                  progress(true, Math.min(uploadedBefore + currentLoaded, total), total)
                },
                controller.signal
              )
              if (finalized) return null

              const etag = directRes?.headers?.etag || directRes?.headers?.ETag
              if (!etag) {
                throw new Error('分块上传成功但未返回ETag，请检查 MinIO CORS 配置')
              }

              await reportUploadedPart({
                upload_id: uploadId,
                part_number: part.part_number,
                etag,
              })
              if (finalized) return null

              progress(true, end, file.size || 1)
            }

            const completeRes = await completeChunkUpload({ upload_id: uploadId })
            return completeRes?.data?.id
          }
          const uploadedId = await uploadChunksWithResume()
          if (finalized || !uploadedId) return
          finishUpload(uploadedId, '上传成功')
          return
        }

        return uploadFile(file, fileStore.currentParentId, (event) => {
          if (finalized) return
          const total = event.total || event.loaded || 0
          if (!total) return
          progress(Boolean(event.lengthComputable || event.total), event.loaded, total)
        }, controller.signal)
          .then((res) => {
            if (finalized) return
            const uploadedId = res?.data?.id
            finishUpload(uploadedId, '上传成功')
          })
      })
      .then((res) => {
        if (!res || finalized) return
      })
      .catch((err) => {
        if (err?.name === 'CanceledError' || err?.code === 'ERR_CANCELED') return
        finalized = true
        error(err.message || '上传失败')
      })

    return {
      abort: () => {
        finalized = true
        controller.abort()
        abort()
      },
    }
  },
  revert: (uniqueFileId, load, error) => {
    const uploadedId = Number(uniqueFileId)
    if (!uploadedId) {
      load()
      return
    }

    deleteFile(uploadedId)
      .then(() => permanentDeleteFile(uploadedId))
      .then(() => {
        load()
        toast.success('已撤销上传')
        Promise.allSettled([
          fileStore.refresh(),
          userStore.fetchUserInfo(),
        ])
      })
      .catch((err) => {
        error(err?.message || '撤销上传失败')
      })
  },
}

function formatTime(str) {
  if (!str) return '-'
  const d = new Date(str)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function handleRowDblClick(row) {
  if (row.is_folder) fileStore.enterFolder(row)
}

function isSelected(id) {
  return selectedIds.value.includes(id)
}

function toggleRowSelection(id, checked) {
  if (checked) {
    if (!selectedIds.value.includes(id)) {
      selectedIds.value = [...selectedIds.value, id]
    }
    return
  }
  selectedIds.value = selectedIds.value.filter((item) => item !== id)
}

function toggleSelectAll(event) {
  selectedIds.value = event.target.checked ? fileStore.files.map((row) => row.id) : []
}

function clearSelection() {
  selectedIds.value = []
}

function closeUpload() {
  uploadPond.value?.removeFiles?.()
  showUploadDialog.value = false
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

function handleDelete(row) {
  deleteTarget.value = row
  deleteTargets.value = [row]
  showDeleteDialog.value = true
}

function handleBatchDelete() {
  if (!selectedRows.value.length) {
    toast.warning('请先勾选要删除的文件')
    return
  }
  deleteTarget.value = null
  deleteTargets.value = [...selectedRows.value]
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!deleteTargets.value.length) return
  deleting.value = true
  try {
    if (deleteTargets.value.length === 1 && deleteTarget.value) {
      await deleteFile(deleteTarget.value.id)
    } else {
      await batchDeleteFiles(deleteTargets.value.map((item) => item.id))
    }
    toast.success('已移入回收站')
    showDeleteDialog.value = false
    deleteTarget.value = null
    deleteTargets.value = []
    clearSelection()
    fileStore.refresh()
    userStore.fetchUserInfo()
  } finally {
    deleting.value = false
  }
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

function extractShareToken(input) {
  const value = input.trim()
  if (!value) return ''
  const shareMatch = value.match(/\/share\/([^/?#]+)/i)
  if (shareMatch?.[1]) return shareMatch[1]
  return value
}

async function handleFetchSharedFile() {
  const shareToken = extractShareToken(shareDownloadForm.link)
  if (!shareToken) {
    shareDownloadError.value = '请输入有效的分享链接'
    return
  }

  shareDownloading.value = true
  shareDownloadError.value = ''
  try {
    const payload = { share_token: shareToken }
    if (shareDownloadForm.key.trim()) payload.key = shareDownloadForm.key.trim()
    const res = await downloadShare(payload)
    const url = res?.data?.URL || res?.data?.url || ''
    const name = res?.data?.FileName || res?.data?.fileName || ''
    if (!url) {
      shareDownloadError.value = '未获取到下载链接'
      return
    }
    sharedDownloadUrl.value = url
    sharedDownloadName.value = name
    sharedDownloadReady.value = true
  } catch (err) {
    shareDownloadError.value = err?.message || '获取文件失败'
  } finally {
    shareDownloading.value = false
  }
}

function downloadSharedFile() {
  if (!sharedDownloadUrl.value) return
  const a = document.createElement('a')
  a.href = sharedDownloadUrl.value
  a.download = sharedDownloadName.value
  a.target = '_blank'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

function resetShareDownloadDialog() {
  shareDownloadForm.link = ''
  shareDownloadForm.key = ''
  sharedDownloadUrl.value = ''
  sharedDownloadName.value = ''
  sharedDownloadReady.value = false
  shareDownloadError.value = ''
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
  if (!shareLink.value) {
    toast.warning('暂无可复制的分享链接')
    return
  }

  const onCopySuccess = () => {
    if (shareForm.key) {
      toast.success(`已复制，提取码 ${shareForm.key}`)
    } else {
      toast.success('链接已复制')
    }
  }

  if (navigator.clipboard?.writeText && window.isSecureContext) {
    navigator.clipboard.writeText(shareLink.value)
      .then(onCopySuccess)
      .catch(() => fallbackCopyShareLink())
    return
  }

  fallbackCopyShareLink()
}

function fallbackCopyShareLink() {
  const textarea = document.createElement('textarea')
  textarea.value = shareLink.value
  textarea.setAttribute('readonly', 'readonly')
  textarea.style.position = 'fixed'
  textarea.style.top = '0'
  textarea.style.left = '0'
  textarea.style.width = '1px'
  textarea.style.height = '1px'
  textarea.style.padding = '0'
  textarea.style.border = '0'
  textarea.style.outline = '0'
  textarea.style.boxShadow = 'none'
  textarea.style.background = 'transparent'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.select()
  textarea.setSelectionRange(0, textarea.value.length)

  try {
    const copied = document.execCommand('copy')
    if (!copied) throw new Error('copy failed')
    if (shareForm.key) {
      toast.success(`已复制，提取码 ${shareForm.key}`)
    } else {
      toast.success('链接已复制')
    }
  } catch {
    toast.error('复制失败，请手动复制输入框中的链接')
  } finally {
    document.body.removeChild(textarea)
  }
}

onMounted(() => {
  fileStore.fetchFiles(0)
})

watch(() => fileStore.files, (files) => {
  const validIds = new Set(files.map((item) => item.id))
  selectedIds.value = selectedIds.value.filter((id) => validIds.has(id))
}, { deep: true })

watch(showDeleteDialog, (open) => {
  if (!open && !deleting.value) {
    deleteTarget.value = null
    deleteTargets.value = []
  }
})

watch(showShareDownloadDialog, (open) => {
  if (!open) resetShareDownloadDialog()
})
</script>

<style scoped>
.selection-checkbox {
  appearance: none;
  -webkit-appearance: none;
  width: 0.875rem;
  height: 0.875rem;
  margin: 0;
  vertical-align: middle;
  background-color: rgb(255 255 255);
  border-radius: 9999px;
  border: 1px solid rgb(231 229 228);
  cursor: pointer;
  display: inline-grid;
  place-content: center;
  transition: border-color 160ms ease, background-color 160ms ease, box-shadow 160ms ease;
}

.selection-checkbox:hover {
  border-color: rgb(214 211 209);
}

.selection-checkbox:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px rgba(214, 211, 209, 0.18);
}

.selection-checkbox::before {
  content: '';
  width: 0.28rem;
  height: 0.28rem;
  border-radius: 9999px;
  transform: scale(0);
  transition: transform 160ms ease;
  background-color: rgb(87 83 78);
}

.selection-checkbox:checked {
  background-color: rgb(245 245 244);
  border-color: rgb(196 181 173);
}

.selection-checkbox:checked::before {
  transform: scale(1);
}

.selection-checkbox:indeterminate {
  background-color: rgb(245 245 244);
  border-color: rgb(214 211 209);
}

.selection-checkbox:indeterminate::before {
  width: 0.38rem;
  height: 0.125rem;
  border-radius: 9999px;
  transform: scale(1);
  background-color: rgb(120 113 108);
}

:deep(.filepond--root) {
  margin: 0;
  font-family: inherit;
}

:deep(.filepond--drop-label) {
  color: rgb(82 82 82);
}

:deep(.filepond--label-action) {
  color: rgb(23 23 23);
  text-decoration-color: rgb(163 163 163);
}

:deep(.filepond--panel-root) {
  background-color: rgb(250 250 250);
  border: 1px solid rgb(229 229 229);
  border-radius: 0.875rem;
}

:deep(.filepond--drip-blob) {
  background-color: rgb(115 115 115);
}

:deep(.filepond--item-panel) {
  background-color: rgb(38 38 38);
}

:deep(.filepond--item > .filepond--file-wrapper),
:deep(.filepond--item > .filepond--panel) {
  transition-duration: 160ms !important;
}

:deep(.filepond--file-action-button) {
  cursor: pointer;
}

:deep([data-filepond-item-state='processing-complete'] .filepond--item-panel) {
  background-color: rgb(34 197 94);
}

:deep([data-filepond-item-state*='invalid'] .filepond--item-panel),
:deep([data-filepond-item-state*='error'] .filepond--item-panel) {
  background-color: rgb(220 38 38);
}
</style>
