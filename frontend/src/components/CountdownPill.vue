<script setup lang="ts">
import { useCountDown } from "@/composables/useCountDown"

const props = defineProps<{
  startTime: string | number | Date
}>()

const { formatted, isStarted } = useCountDown(() => props.startTime)
</script>

<template>
  <span class="border border-[#1C1C1C]/20 px-3 py-1 text-xs tracking-wide uppercase">
    {{ isStarted ? "进行中" : formatted }}
  </span>
</template>
