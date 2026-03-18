<script setup lang="ts">
import { RouterView } from "vue-router"
import { Toaster } from "vue-sonner"
</script>

<template>
  <div class="min-h-screen bg-[#F9F8F6] font-sans text-[#1C1C1C]">
    <RouterView />
    <Toaster position="top-right" theme="light" :rich-colors="false" />
  </div>
</template>
