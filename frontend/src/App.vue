<script setup lang="ts">
import { RouterView } from "vue-router"
import { Toaster } from "vue-sonner"
</script>

<template>
  <div class="min-h-screen bg-obsidian-bg text-white">
    <RouterView />
    <Toaster position="top-right" rich-colors />
  </div>
</template>
