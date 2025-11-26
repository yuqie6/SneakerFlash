export type OrderStatus = 0 | 1 | 2

export interface Order {
  id: number
  user_id: number
  product_id: number
  order_num: string
  status: OrderStatus
  created_at?: string
  updated_at?: string
}

import type { Payment } from "./payment"
import type { Coupon } from "./coupon"

export interface OrderWithPayment {
  order: Order
  payment?: Payment
  coupon?: Coupon
}
