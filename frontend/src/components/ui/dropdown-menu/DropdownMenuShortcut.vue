<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})

const classes = computed(() => cn('ml-auto text-xs tracking-widest text-neutral-500', props.class))
</script>

<template>
  <span :class="classes">
    <slot />
  </span>
</template>
