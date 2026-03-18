<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <div
    :class="
      cn(
        'border border-[#1C1C1C]/10 bg-white text-[#1C1C1C] transition-colors hover:border-[#1C1C1C]',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
