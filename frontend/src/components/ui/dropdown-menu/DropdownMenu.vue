<script setup>
import { DropdownMenuRoot } from 'radix-vue'

defineProps({
  open: {
    type: Boolean,
    default: undefined,
  },
  modal: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['update:open'])
</script>

<template>
  <DropdownMenuRoot :open="open" :modal="modal" @update:open="emit('update:open', $event)">
    <slot />
  </DropdownMenuRoot>
</template>
