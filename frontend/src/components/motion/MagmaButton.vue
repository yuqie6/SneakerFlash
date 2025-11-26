<template>
  <button
    class="relative group overflow-hidden rounded-full bg-magma-gradient px-8 py-4 font-bold text-white shadow-[0_0_20px_-5px_rgba(249,115,22,0.4)] transition-all hover:scale-[1.02] active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
    :disabled="disabled || loading"
  >
    <div
      class="absolute inset-0 -translate-x-full bg-gradient-to-r from-transparent via-white/25 to-transparent group-hover:animate-shimmer z-10"
    ></div>
    <span class="relative z-20 flex items-center justify-center gap-2">
      <svg
        v-if="loading"
        class="h-5 w-5 animate-spin"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
      </svg>
      <slot />
    </span>
  </button>
</template>

<script setup lang="ts">
defineProps<{ loading?: boolean; disabled?: boolean }>()
</script>
