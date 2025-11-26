import { ref } from "vue"
import { useRouter } from "vue-router"
import { toast } from "vue-sonner"
import api from "@/lib/api"
import { useUserStore } from "@/stores/userStore"

export type SeckillStatus = "idle" | "loading" | "success" | "failed"

export type SeckillResult = {
  order_num: string
  order_id: number
  payment_id?: string
}

export function useSeckill() {
  const status = ref<SeckillStatus>("idle")
  const resultMsg = ref("")
  const result = ref<SeckillResult | null>(null)
  const router = useRouter()
  const userStore = useUserStore()

  const executeSeckill = async (productId: number) => {
    const token = userStore.accessToken || localStorage.getItem("access_token")
    if (!token) {
      toast.error("请先登录")
      router.push({ name: "login" })
      return null
    }

    status.value = "loading"
    result.value = null
    try {
      const res = await api.post<SeckillResult, SeckillResult>("/seckill", { product_id: productId })
      result.value = res
      status.value = "success"
      resultMsg.value = `抢购成功！订单号: ${res.order_num || ""}`
      toast.success("GOT 'EM!", { description: "已生成支付单，正在跳转" })
      return res
    } catch (err: any) {
      status.value = "failed"
      resultMsg.value = err?.message || "抢购失败"
      toast.error("抢购失败", { description: resultMsg.value })
      return null
    }
  }

  return { status, resultMsg, result, executeSeckill }
}
