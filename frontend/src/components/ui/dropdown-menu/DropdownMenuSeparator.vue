<script setup>
import { DropdownMenuSeparator } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})
</script>

<template>
  <DropdownMenuSeparator :class="cn('-mx-1 my-1 h-px bg-neutral-100', props.class)" />
</template>