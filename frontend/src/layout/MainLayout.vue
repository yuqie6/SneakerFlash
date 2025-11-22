<template>
  <div class="flex min-h-screen flex-col bg-obsidian-bg text-white">
    <header class="sticky top-0 z-30 border-b border-obsidian-border/70 bg-obsidian-bg/70 backdrop-blur-xl">
      <div class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <RouterLink to="/" class="text-lg font-semibold tracking-tight">
          SneakerFlash
        </RouterLink>
        <nav class="flex items-center gap-4 text-sm text-white/80">
          <RouterLink class="transition hover:text-white" to="/">首页</RouterLink>
          <RouterLink class="transition hover:text-white" to="/login">登录</RouterLink>
        </nav>
      </div>
    </header>
    <main class="flex-1">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from "vue-router"
</script>
