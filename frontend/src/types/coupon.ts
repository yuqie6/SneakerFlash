export type CouponType = "full_cut" | "discount"
export type CouponStatus = "available" | "used" | "expired"

export interface Coupon {
  id: number
  coupon_id: number
  type: CouponType
  title: string
  description: string
  amount_cents: number
  discount_rate: number
  min_spend_cents: number
  status: CouponStatus
  valid_from: string
  valid_to: string
  obtained_from: string
}
