<script setup>
import { computed } from 'vue'
import { cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const props = defineProps({
  variant: {
    type: String,
    default: 'default',
  },
  class: {
    type: String,
    default: '',
  },
})

const badgeVariants = cva(
  'inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium transition-colors',
  {
    variants: {
      variant: {
        default: 'border-neutral-900 bg-neutral-900 text-white',
        secondary: 'border-border bg-neutral-100 text-neutral-700',
        outline: 'border-border bg-white text-neutral-700',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

const classes = computed(() => cn(badgeVariants({ variant: props.variant }), props.class))
</script>

<template>
  <span :class="classes">
    <slot />
  </span>
</template>
