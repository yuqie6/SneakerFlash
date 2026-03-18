<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center border px-2.5 py-0.5 text-xs tracking-[0.2em] uppercase transition-colors focus:outline-none",
  {
    variants: {
      variant: {
        default: "border-[#1C1C1C] bg-[#1C1C1C] text-white",
        secondary: "border-[#1C1C1C]/10 bg-[#1C1C1C]/5 text-[#1C1C1C]",
        destructive: "border-[#1C1C1C]/20 bg-transparent text-[#1C1C1C]/70",
        outline: "border-[#1C1C1C]/20 text-[#1C1C1C]",
        success: "border-[#1C1C1C]/20 bg-[#1C1C1C]/5 text-[#1C1C1C]/70",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

type BadgeVariants = VariantProps<typeof badgeVariants>

interface Props {
  variant?: BadgeVariants["variant"]
  class?: HTMLAttributes["class"]
}

const props = withDefaults(defineProps<Props>(), {
  variant: "default",
})
</script>

<template>
  <div :class="cn(badgeVariants({ variant }), props.class)">
    <slot />
  </div>
</template>
