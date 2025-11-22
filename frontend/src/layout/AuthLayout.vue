<template>
  <div class="flex min-h-screen items-center justify-center bg-obsidian-bg text-white">
    <div class="absolute inset-0 bg-gradient-to-b from-magma-glow/10 via-transparent to-black pointer-events-none"></div>
    <div class="relative z-10 w-full max-w-md rounded-3xl border border-white/10 bg-obsidian-card/70 p-8 shadow-2xl backdrop-blur-xl">
      <RouterLink to="/" class="mb-6 block text-center text-xl font-semibold">SneakerFlash</RouterLink>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from "vue-router"
</script>
