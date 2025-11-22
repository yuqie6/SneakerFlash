<script setup lang="ts">
import { useCountDown } from "@/composables/useCountDown"

const props = defineProps<{
  startTime: string | number | Date
}>()

const { formatted, isStarted } = useCountDown(() => props.startTime)
</script>

<template>
  <span class="rounded-full border border-obsidian-border px-3 py-1 text-xs">
    {{ isStarted ? "进行中" : formatted }}
  </span>
</template>
