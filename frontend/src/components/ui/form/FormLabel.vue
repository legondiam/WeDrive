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

const classes = computed(() => cn('text-sm font-medium leading-none text-foreground', props.class))
</script>

<template>
  <label :class="classes" v-bind="$attrs">
    <slot />
  </label>
</template>
