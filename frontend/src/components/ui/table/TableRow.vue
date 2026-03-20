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
  cn('border-b transition-colors hover:bg-neutral-50 data-[state=selected]:bg-neutral-100', props.class)
)
</script>

<template>
  <tr :class="classes">
    <slot />
  </tr>
</template>
