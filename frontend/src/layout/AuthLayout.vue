<template>
  <div class="flex min-h-screen items-center justify-center bg-[#F9F8F6] text-[#1C1C1C]">
    <div class="relative z-10 w-full max-w-md border border-[#1C1C1C]/10 bg-white p-8">
      <RouterLink to="/" class="mb-6 block text-center font-serif text-xl tracking-tight">SneakerFlash</RouterLink>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from "vue-router"
</script>
