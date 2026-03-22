<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'
defineOptions({ inheritAttrs: false })

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: '',
  },
  type: {
    type: String,
    default: 'text',
  },
  placeholder: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  class: {
    type: String,
    default: '',
  },
  min: {
    type: Number,
    default: undefined,
  },
  max: {
    type: Number,
    default: undefined,
  },
  readonly: {
    type: Boolean,
    default: false,
  },
  maxlength: {
    type: [Number, String],
    default: undefined,
  },
})

const emit = defineEmits(['update:modelValue', 'keyup', 'change'])

const classes = computed(() =>
  cn(
    'flex h-10 w-full rounded-md border border-border bg-white px-3 py-2 text-sm text-foreground shadow-sm transition-[border-color,background-color,box-shadow] duration-200 placeholder:text-neutral-400 focus-visible:outline-none focus-visible:border-neutral-400 focus-visible:bg-white focus-visible:shadow-[0_0_0_3px_rgba(23,23,23,0.06),0_1px_2px_rgba(0,0,0,0.04)] disabled:cursor-not-allowed disabled:opacity-50',
    props.class
  )
)

function onInput(event) {
  emit('update:modelValue', event.target.value)
}
</script>

<template>
  <input
    :value="modelValue"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :min="min"
    :max="max"
    :readonly="readonly"
    :maxlength="maxlength"
    :class="classes"
    v-bind="$attrs"
    @input="onInput"
    @keyup="emit('keyup', $event)"
    @change="emit('change', $event)"
  />
</template>
