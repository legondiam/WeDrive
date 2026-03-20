<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})

const classes = computed(() => cn('text-[12px] font-medium leading-[1.6] text-red-500', props.class))
</script>

<template>
  <p :class="classes" role="alert">
    <slot />
  </p>
</template>
