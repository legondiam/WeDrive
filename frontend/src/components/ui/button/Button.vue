<script setup>
import { computed } from 'vue'
import { cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const props = defineProps({
  variant: {
    type: String,
    default: 'default',
  },
  size: {
    type: String,
    default: 'default',
  },
  class: {
    type: String,
    default: '',
  },
  type: {
    type: String,
    default: 'button',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md border text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'border-black bg-black text-white hover:bg-neutral-700 hover:border-neutral-700 hover:shadow-md',
        outline: 'border-border bg-white text-foreground hover:border-neutral-300 hover:bg-neutral-50',
        ghost: 'border-transparent bg-transparent text-foreground hover:bg-neutral-100',
        destructive: 'border-neutral-700 bg-neutral-700 text-white hover:bg-neutral-900 hover:border-neutral-900',
        secondary: 'border-border bg-neutral-100 text-foreground hover:bg-neutral-200',
        link: 'border-transparent p-0 text-foreground underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-10 px-4 py-2',
        sm: 'h-9 px-3',
        lg: 'h-11 px-6',
        icon: 'h-9 w-9 p-0',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
)

const classes = computed(() => cn(buttonVariants({ variant: props.variant, size: props.size }), props.class))
</script>

<template>
  <button :type="type" :disabled="disabled" :class="classes">
    <slot />
  </button>
</template>
