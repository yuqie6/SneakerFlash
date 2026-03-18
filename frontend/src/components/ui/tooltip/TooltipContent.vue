<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { TooltipContent, TooltipPortal, type TooltipContentProps } from "reka-ui"
import { cn } from "@/lib/utils"

interface Props extends TooltipContentProps {
  class?: HTMLAttributes["class"]
}

const props = withDefaults(defineProps<Props>(), {
  sideOffset: 4,
})
</script>

<template>
  <TooltipPortal>
    <TooltipContent
      v-bind="props"
      :class="cn(
        'z-50 overflow-hidden border border-[#1C1C1C]/10 bg-white px-3 py-1.5 text-sm text-[#1C1C1C] animate-in fade-in-0 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
        props.class
      )"
    >
      <slot />
    </TooltipContent>
  </TooltipPortal>
</template>
