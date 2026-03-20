<script setup>
import { DropdownMenuSubContent } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})
</script>

<template>
  <DropdownMenuSubContent
    :class="cn('z-50 min-w-[8rem] overflow-hidden rounded-md border border-border bg-white p-1 shadow-md data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95', props.class)"
  >
    <slot />
  </DropdownMenuSubContent>
</template>
