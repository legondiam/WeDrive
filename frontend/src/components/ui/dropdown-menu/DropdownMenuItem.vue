<script setup>
import { DropdownMenuItem } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})
</script>

<template>
  <DropdownMenuItem
    :class="cn('relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-neutral-100 focus:text-neutral-900 data-[disabled]:pointer-events-none data-[disabled]:opacity-50', props.class)"
  >
    <slot />
  </DropdownMenuItem>
</template>