<template>
  <div
    ref="cardRef"
    class="relative overflow-hidden border border-[#1C1C1C]/10 bg-white p-4 transition-colors duration-200"
    :style="cardStyle"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useMouseInElement } from "@vueuse/core"

const cardRef = ref<HTMLElement | null>(null)
const { elementY, elementHeight, isOutside } = useMouseInElement(cardRef)

const cardStyle = computed(() => {
  if (isOutside.value) {
    return {
      transform: "translateY(0)",
    }
  }

  const verticalOffset = Math.min(
    4,
    Math.max(
      0,
      ((elementHeight.value / 2 - Math.abs(elementY.value - elementHeight.value / 2)) / elementHeight.value) * 6,
    ),
  )

  return {
    transform: `translateY(-${verticalOffset}px)`,
  }
})
</script>
