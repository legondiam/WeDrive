<template>
  <component
    :is="as"
    :to="to"
    :href="href"
    :class="cn(
      'group bg-background relative w-auto cursor-pointer overflow-hidden rounded-full border p-2 px-6 text-center font-semibold [contain:paint]',
      className,
    )"
  >
    <span
      aria-hidden="true"
      class="absolute inset-0 rounded-full bg-primary transition-[clip-path] duration-300 ease-out [clip-path:circle(0.25rem_at_1.25rem_50%)] group-hover:[clip-path:circle(150%_at_1.25rem_50%)]"
    />
    <div class="flex items-center justify-center">
      <span class="ihb-text inline-block pl-4 transform-gpu transition-all duration-300 will-change-transform group-hover:translate-x-12 group-hover:opacity-0">
        <slot>{{ text }}</slot>
      </span>
    </div>
    <div class="ihb-hover text-primary-foreground absolute top-0 z-10 flex h-full w-full translate-x-12 transform-gpu items-center justify-center gap-2 opacity-0 transition-all duration-300 will-change-transform group-hover:-translate-x-5 group-hover:opacity-100">
      <span>
        <slot>{{ text }}</slot>
      </span>
      <ArrowRight class="transform-gpu" />
    </div>
  </component>
</template>

<script setup>
import { computed } from 'vue'
import { ArrowRight } from 'lucide-vue-next'

import { cn } from '@/lib/utils'

const props = defineProps({
  as: {
    type: [String, Object],
    default: 'button',
  },
  text: {
    type: String,
    default: 'Button',
  },
  class: {
    type: String,
    default: '',
  },
  to: {
    type: [String, Object],
    default: undefined,
  },
  href: {
    type: String,
    default: undefined,
  },
})

const className = computed(() => props.class)
</script>
