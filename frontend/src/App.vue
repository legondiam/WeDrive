<template>
  <router-view />
  <Toaster
    position="top-center"
    rich-colors
    close-button
    :duration="1800"
    :toast-options="{ class: 'min-h-0 w-auto min-w-0 max-w-[280px] rounded-xl px-3 py-2' }"
  />
</template>

<script setup>
import { Toaster } from 'vue-sonner'
</script>
