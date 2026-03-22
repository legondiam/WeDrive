<template>
  <component
    :is="as"
    :to="to"
    :href="href"
    :class="cn(
      'group bg-background relative isolate w-auto cursor-pointer overflow-hidden rounded-full border p-2 px-6 text-center font-semibold [contain:layout_paint_style]',
      className,
    )"
  >
    <svg
      aria-hidden="true"
      viewBox="0 0 240 240"
      class="pointer-events-none absolute left-[1.05rem] top-1/2 z-0 h-60 w-60 -translate-x-1/2 -translate-y-1/2 origin-center scale-[0.032] transform-gpu transition-transform duration-300 ease-out will-change-transform [backface-visibility:hidden] group-hover:scale-100"
      preserveAspectRatio="xMidYMid meet"
      shape-rendering="geometricPrecision"
    >
      <circle cx="120" cy="120" r="120" class="fill-zinc-800" />
    </svg>
    <div class="relative z-10 flex items-center justify-center">
      <span class="ihb-text inline-block translate-x-[0.08rem] transform-gpu transition-[transform,opacity] duration-300 ease-out will-change-transform group-hover:translate-x-12 group-hover:opacity-0 [backface-visibility:hidden]">
        <slot>{{ text }}</slot>
      </span>
    </div>
    <div class="ihb-hover text-primary-foreground pointer-events-none absolute top-0 z-10 flex h-full w-full translate-x-12 transform-gpu items-center justify-center gap-1.5 opacity-0 transition-[transform,opacity] duration-300 ease-out will-change-transform group-hover:-translate-x-5 group-hover:opacity-100 [backface-visibility:hidden]">
      <span>
        <slot>{{ text }}</slot>
      </span>
      <ArrowRight class="transform-gpu [backface-visibility:hidden]" />
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
