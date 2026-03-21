<script setup>
import { DialogClose, DialogContent, DialogOverlay, DialogPortal } from 'radix-vue'
import { X } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
  showClose: {
    type: Boolean,
    default: true,
  },
})
</script>

<template>
  <DialogPortal>
    <DialogOverlay class="fixed inset-0 z-50 bg-black/30" />
    <DialogContent
      :class="cn('fixed left-1/2 top-1/2 z-50 w-[calc(100%-2rem)] max-w-lg -translate-x-1/2 -translate-y-1/2 rounded-md border border-border bg-white p-4 shadow-md focus:outline-none', props.class)"
    >
      <slot />
      <DialogClose
        v-if="showClose"
        class="absolute right-3 top-3 inline-flex h-7 w-7 items-center justify-center rounded-md border border-border bg-white text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-900"
        aria-label="关闭"
      >
        <X class="h-4 w-4" />
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>
