<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  value: {
    type: Number,
    default: 0,
  },
  class: {
    type: String,
    default: '',
  },
})

const normalized = computed(() => Math.min(100, Math.max(0, Number(props.value) || 0)))
const classes = computed(() => cn('relative h-2 w-full overflow-hidden rounded-full bg-neutral-100', props.class))
</script>

<template>
  <div :class="classes">
    <div
      class="h-full rounded-full bg-neutral-900 transition-all duration-300"
      :style="{ width: `${normalized}%` }"
    />
  </div>
</template>
