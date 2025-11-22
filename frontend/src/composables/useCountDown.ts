import { computed, onBeforeUnmount, onMounted, ref, toValue, watch, type MaybeRef } from "vue"

const toTimestamp = (input: Date | string | number) => new Date(input).getTime()

export function useCountDown(target: MaybeRef<Date | string | number>) {
  const targetTime = ref<number>(toTimestamp(toValue(target)))
  const remaining = ref(0)
  let timer: number | undefined

  const tick = () => {
    const now = Date.now()
    remaining.value = Math.max(0, Math.floor((targetTime.value - now) / 1000))
  }

  const start = () => {
    tick()
    stop()
    timer = window.setInterval(tick, 1000)
  }

  const stop = () => {
    if (timer) {
      clearInterval(timer)
      timer = undefined
    }
  }

  const formatted = computed(() => {
    const minutes = Math.floor(remaining.value / 60)
    const seconds = remaining.value % 60
    return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`
  })

  const isStarted = computed(() => remaining.value === 0)

  watch(() => toValue(target), (val) => {
    targetTime.value = toTimestamp(val)
    start()
  })

  onMounted(start)
  onBeforeUnmount(stop)

  return { remaining, formatted, isStarted, start, stop }
}
