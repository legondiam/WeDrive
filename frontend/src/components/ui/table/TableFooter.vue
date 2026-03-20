<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  class: {
    type: String,
    default: '',
  },
})

const classes = computed(() => cn('border-t bg-neutral-50/50 font-medium [&>tr]:last:border-b-0', props.class))
</script>

<template>
  <tfoot :class="classes">
    <slot />
  </tfoot>
</template>
