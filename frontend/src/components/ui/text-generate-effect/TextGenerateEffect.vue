<template>
  <component :is="as" :class="cn('inline-block whitespace-pre-wrap', className)">
    <span
      v-for="(segment, index) in segments"
      :key="`${segment}-${index}`"
      :class="[
        'transition-all ease-out',
        revealedCount > index ? 'translate-y-0 opacity-100 blur-0' : 'translate-y-2 opacity-0 blur-[10px]',
      ]"
      :style="{ transitionDuration: `${duration}ms`, transitionDelay: `${index * stagger}ms` }"
    >
      {{ segment }}
    </span>
  </component>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import { cn } from '@/lib/utils'

const props = defineProps({
  text: {
    type: String,
    required: true,
  },
  as: {
    type: String,
    default: 'div',
  },
  duration: {
    type: Number,
    default: 500,
  },
  stagger: {
    type: Number,
    default: 45,
  },
  startDelay: {
    type: Number,
    default: 0,
  },
  class: {
    type: String,
    default: '',
  },
})

const className = computed(() => props.class)
const segments = computed(() => Array.from(props.text))
const revealedCount = ref(0)

let startTimer = null
let intervalTimer = null

function clearTimers() {
  if (startTimer) {
    clearTimeout(startTimer)
    startTimer = null
  }

  if (intervalTimer) {
    clearInterval(intervalTimer)
    intervalTimer = null
  }
}

function startAnimation() {
  clearTimers()
  revealedCount.value = 0

  startTimer = setTimeout(() => {
    intervalTimer = setInterval(() => {
      if (revealedCount.value >= segments.value.length) {
        clearTimers()
        revealedCount.value = segments.value.length
        return
      }

      revealedCount.value += 1
    }, props.stagger)
  }, props.startDelay)
}

watch(() => props.text, startAnimation)

onMounted(startAnimation)
onBeforeUnmount(clearTimers)
</script>
