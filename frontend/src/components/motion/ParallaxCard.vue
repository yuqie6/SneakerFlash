<template>
  <div
    ref="cardRef"
    class="relative overflow-hidden rounded-2xl border border-obsidian-border/60 bg-obsidian-card/80 p-4 transition-transform duration-300 will-change-transform"
    :style="cardStyle"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useMouseInElement } from "@vueuse/core"

const cardRef = ref<HTMLElement | null>(null)
const { elementX, elementY, elementWidth, elementHeight, isOutside } = useMouseInElement(cardRef)

const cardStyle = computed(() => {
  if (isOutside.value) {
    return {
      transform: "rotateX(0deg) rotateY(0deg) translateZ(0)",
    }
  }

  const rotateX = ((elementY.value - elementHeight.value / 2) / elementHeight.value) * -10
  const rotateY = ((elementX.value - elementWidth.value / 2) / elementWidth.value) * 10

  return {
    transform: `perspective(900px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) translateZ(12px)`,
  }
})
</script>
