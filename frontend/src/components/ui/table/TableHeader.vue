<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})

const classes = computed(() => cn('[&_tr]:border-b', props.class))
</script>

<template>
  <thead :class="classes">
    <slot />
  </thead>
</template>
