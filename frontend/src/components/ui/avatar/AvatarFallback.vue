<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { AvatarFallback } from "reka-ui"
import { cn } from "@/lib/utils"

interface Props {
  class?: HTMLAttributes["class"]
}

const props = defineProps<Props>()
</script>

<template>
  <AvatarFallback
    :class="cn('flex h-full w-full items-center justify-center rounded-full bg-obsidian-card text-white', props.class)"
  >
    <slot />
  </AvatarFallback>
</template>
