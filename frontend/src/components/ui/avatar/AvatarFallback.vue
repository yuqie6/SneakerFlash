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
    :class="cn('flex h-full w-full items-center justify-center bg-[#1C1C1C]/5 text-[#1C1C1C]', props.class)"
  >
    <slot />
  </AvatarFallback>
</template>
