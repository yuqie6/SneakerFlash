<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { AvatarImage } from "reka-ui"
import { cn } from "@/lib/utils"

interface Props {
  src?: string
  alt?: string
  class?: HTMLAttributes["class"]
}

const props = defineProps<Props>()
</script>

<template>
  <AvatarImage
    :src="src ?? ''"
    :alt="alt ?? ''"
    :class="cn('aspect-square h-full w-full object-cover', props.class)"
  />
</template>
