<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})

const classes = computed(() => cn('[&_tr:last-child]:border-0', props.class))
</script>

<template>
  <tbody :class="classes">
    <slot />
  </tbody>
</template>
