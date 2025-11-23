import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatPrice(price: number) {
  const n = Number(price)
  if (!Number.isFinite(n)) return "¥0.00"
  return n.toLocaleString("zh-CN", { style: "currency", currency: "CNY" })
}
