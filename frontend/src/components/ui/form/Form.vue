<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'
defineOptions({ inheritAttrs: false })

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})

const classes = computed(() => cn('space-y-4', props.class))
</script>

<template>
  <form :class="classes" v-bind="$attrs">
    <slot />
  </form>
</template>
