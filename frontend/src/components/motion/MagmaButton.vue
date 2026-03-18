<template>
  <button
    class="border border-[#1C1C1C] bg-[#1C1C1C] px-6 py-3 text-sm tracking-wide text-white transition-colors hover:bg-[#1C1C1C]/90 active:scale-95 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:bg-[#1C1C1C]"
    :disabled="disabled || loading"
  >
    <span class="flex items-center justify-center gap-2">
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
