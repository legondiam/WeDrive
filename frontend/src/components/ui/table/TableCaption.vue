<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})

const classes = computed(() => cn('mt-4 text-sm text-neutral-500', props.class))
</script>

<template>
  <caption :class="classes">
    <slot />
  </caption>
</template>
