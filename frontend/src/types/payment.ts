export type PaymentStatus = "pending" | "paid" | "failed" | "refunded"

export interface Payment {
  id: number
  order_id: number
  payment_id: string
  amount_cents: number
  status: PaymentStatus
  notify_data?: string
  created_at?: string
  updated_at?: string
}
