import { ref } from "vue"
import { useRouter } from "vue-router"
import { toast } from "vue-sonner"
import api from "@/lib/api"
import { useUserStore } from "@/stores/userStore"

export type SeckillStatus = "idle" | "loading" | "success" | "failed"

export function useSeckill() {
  const status = ref<SeckillStatus>("idle")
  const resultMsg = ref("")
  const router = useRouter()
  const userStore = useUserStore()

  const executeSeckill = async (productId: number) => {
    const token = userStore.accessToken || localStorage.getItem("access_token")
    if (!token) {
      toast.error("请先登录")
      router.push({ name: "login" })
      return
    }

    status.value = "loading"
    try {
      const res: any = await api.post("/seckill", { product_id: productId })
      status.value = "success"
      resultMsg.value = `抢购成功！订单号: ${res.order_num || res?.data?.order_num || ""}`
      toast.success("GOT 'EM!", { description: "恭喜，您已成功抢购！" })
    } catch (err: any) {
      status.value = "failed"
      resultMsg.value = err?.message || "抢购失败"
      toast.error("抢购失败", { description: resultMsg.value })
    }
  }

  return { status, resultMsg, executeSeckill }
}
