export interface AdminStats {
  total_users: number
  total_orders: number
  total_revenue_cents: number
  total_products: number
  pending_orders: number
}

export interface CouponTemplate {
  id: number
  type: "full_cut" | "discount"
  title: string
  description: string
  amount_cents: number
  discount_rate: number
  min_spend_cents: number
  valid_from: string
  valid_to: string
  purchasable: boolean
  price_cents: number
  status: string
}

export interface AdminUser {
  id: number
  username: string
  balance: number
  avatar?: string
  total_spent_cents: number
  growth_level: number
  role: string
  created_at: string
}

export interface RiskList {
  ip: string[]
  user: string[]
}
