<script setup>
import { DropdownMenuSubTrigger } from 'radix-vue'
import { ChevronRight } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
  inset: {
    type: Boolean,
    default: false,
  },
})
</script>

<template>
  <DropdownMenuSubTrigger
    :class="cn('flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none focus:bg-neutral-100 data-[state=open]:bg-neutral-100', inset && 'pl-8', props.class)"
  >
    <slot />
    <ChevronRight class="ml-auto h-4 w-4" />
  </DropdownMenuSubTrigger>
</template>
