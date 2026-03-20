<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})

const classes = computed(() =>
  cn('h-10 px-4 text-left align-middle text-[12px] font-medium text-neutral-600 [&:has([role=checkbox])]:pr-0', props.class)
)
</script>

<template>
  <th :class="classes">
    <slot />
  </th>
</template>
