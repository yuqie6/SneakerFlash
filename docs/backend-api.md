# SneakerFlash 后端接口文档（当前实现）

- 基础地址：业务接口 `http://localhost:8000/api/v1`
- 探针接口：`http://localhost:8000/health`、`http://localhost:8000/ready`、`http://localhost:8000/metrics`
- 鉴权：受保护接口需在 Header 携带 `Authorization: Bearer <access_token>`
- 统一返回：`{ "code": number, "msg": string, "data"?: any }`；`code=200` 表示业务成功

## 系统
- `GET /health`（根路径，非 `/api/v1`）
  存活检查。成功：`data={"status":"ok","service":"SneakerFlash","timestamp":RFC3339}`。
- `GET /ready`（根路径，非 `/api/v1`）
  就绪检查，覆盖 MySQL、Redis、Kafka Producer。未就绪返回 HTTP `503`。
- `GET /metrics`（根路径，非 `/api/v1`）
  Prometheus 文本指标。

## 认证
- `POST /register`
  Body：`{ "user_name": string, "user_password": string }`
  成功：`data={"message":"注册成功"}`；重名返回 `code=10001`。
- `POST /login`
  Body 同上。
  成功：`data={ "access_token", "refresh_token", "expires_in" }`。
- `POST /refresh`
  Body：`{ "refresh_token": string }`
  成功：`data={ "access_token", "expires_in" }`。

## 用户与上传
- `GET /profile`（鉴权）
  成功：`data=User`，包含 `total_spent_cents`、`growth_level`、`role`。
- `PUT /profile`（鉴权）
  Body：`{ "user_name"?: string, "avatar"?: string }`；至少传一项。
- `POST /upload`（鉴权）
  `multipart/form-data`，字段 `file`；成功：`data={ "url": string }`。

## 商品
- `GET /products?page=1&page_size=10`
  成功：`data={ list: Product[], total, page, page_size }`。
- `GET /product/:id`
  成功：`data=Product`；不存在返回 `404` + `code=20001`。
- `POST /products`（鉴权）
  Body：`{ name, price, stock, start_time, end_time?, image? }`
  - `start_time` 必须晚于当前时间
  - `end_time` 可选，若传入必须晚于 `start_time`
- `PUT /products/:id`（鉴权，仅发布者）
  Body：同上，支持部分更新；`end_time=""` 表示清空结束时间。
- `DELETE /products/:id`（鉴权，仅发布者）
  成功：`data={ "id": number }`。
- `GET /products/mine?page=1&page_size=10`（鉴权）
  成功：`data={ list: Product[], total, page, page_size }`。

## 秒杀
- `POST /seckill`（鉴权）
  Body：`{ "product_id": number }`
  成功：`data={ "order_num": string, "payment_id": string, "status": "pending"|"ready" }`。
  常见业务码：`30001` 售罄、`30002` 重复下单、`30003` 请求过于频繁。
  未开始/已结束当前返回 `400 + code=400`；系统繁忙当前返回 `503 + code=500`。

## 订单与支付
- `GET /orders?page=1&page_size=10&status=0|1|2`（鉴权）
  成功：`data={ list: Order[], total, page, page_size }`。
- `GET /orders/:id`（鉴权，仅本人）
  成功：`data={ order: Order, payment?: Payment, coupon?: MyCoupon }`。
- `GET /orders/poll/:order_num`（鉴权）
  轮询异步建单结果：
  - `pending`：`{ status, order_num, payment_id? }`
  - `ready`：`{ status, order_num, payment_id, order }`
  - `failed`：`{ status, order_num, message }`
- `POST /orders/:id/apply-coupon`（鉴权，仅本人）
  Body：`{ "coupon_id": number | null }`
  `coupon_id` 为空时表示移除已用优惠券。
  成功：`data={ order, payment?, coupon? }`。
- `POST /payment/callback`
  Body：`{ "payment_id": string, "status": "paid"|"failed"|"refunded", "notify_data"?: string }`
  成功：`data={ order, payment, coupon? }`；支付单不存在返回 `404`。
  `notify_data` 支持持久化完整回调负载，不再受 20 字符限制。

## VIP 与优惠券
- `GET /vip/profile`（鉴权）
  成功：`data={ total_spent_cents, growth_level, paid_level, paid_expired_at, effective_level }`。
- `POST /vip/purchase`（鉴权）
  Body：`{ "plan_id": 1|2 }`
  当前套餐：
  - `1`：L3，30 天
  - `2`：L4，90 天
  当前为模拟购买成功，直接生效并尝试发放当月 VIP 券。
- `GET /coupons/mine?status=available|used|expired&page=1&page_size=20`（鉴权）
  成功：`data={ list: MyCoupon[], total, page, page_size }`。
- `POST /coupons/purchase`（鉴权）
  Body：`{ "coupon_id": number }`
  成功：`data=MyCoupon`；若模板不可购买或已失效会返回业务错误。

## 管理后台
- 鉴权要求：所有 `/admin/*` 接口都需要管理员 `access_token`；普通用户会收到 HTTP `403` + `msg="需要管理员权限"`
- `GET /admin/stats`
  成功：`data={ total_users, total_orders, total_revenue_cents, total_products, pending_orders }`。
- `GET /admin/users?page=1&page_size=20`
  成功：`data={ list: User[], total, page, page_size }`。
- `GET /admin/orders?page=1&page_size=20&status=0|1|2`
  成功：`data={ list: Order[], total, page, page_size }`。
- `GET /admin/products?page=1&page_size=20`
  成功：`data={ list: Product[], total, page, page_size }`。
- `GET /admin/coupons?page=1&page_size=20`
  成功：`data={ list: Coupon[], total, page, page_size }`。
- `POST /admin/coupons`
  Body：`{ type, title, description, amount_cents, discount_rate, min_spend_cents, valid_from, valid_to, purchasable, price_cents, status }`
  - `type`：`full_cut | discount`
  - `status`：`active | inactive`
  - 时间支持 `RFC3339`、`YYYY-MM-DD HH:mm[:ss]`、`YYYY-MM-DDTHH:mm[:ss]`
- `PUT /admin/coupons/:id`
  Body 同上，支持部分字段更新。
- `DELETE /admin/coupons/:id`
  成功：`data={ "message": "ok" }`。
- `GET /admin/risk/blacklist` / `GET /admin/risk/graylist`
  成功：`data={ ip: string[], user: string[] }`。
- `POST /admin/risk/blacklist` / `POST /admin/risk/graylist`
  Body：`{ "type": "ip"|"user", "value": string }`。
- `DELETE /admin/risk/blacklist` / `DELETE /admin/risk/graylist`
  Body 同上；成功：`data={ "message": "ok" }`。

## 数据模型（核心字段）
- `User`：`id`, `username`, `balance`, `avatar`, `total_spent_cents`, `growth_level`, `role`, `created_at`, `updated_at`
- `Product`：`id`, `user_id`, `name`, `price`, `stock`, `start_time`, `end_time`, `image`, `created_at`, `updated_at`
- `Order`：`id`, `user_id`, `product_id`, `order_num`, `status`, `created_at`, `updated_at`
- `Payment`：`id`, `order_id`, `payment_id`, `amount_cents`, `status`, `notify_data`, `created_at`, `updated_at`
- `Coupon`：`id`, `type`, `title`, `description`, `amount_cents`, `discount_rate`, `min_spend_cents`, `valid_from`, `valid_to`, `purchasable`, `price_cents`, `status`
- `MyCoupon`：`id`, `coupon_id`, `type`, `title`, `description`, `amount_cents`, `discount_rate`, `min_spend_cents`, `status`, `valid_from`, `valid_to`, `obtained_from`

## 风控与限流
- 开关：`risk.enable`
- 接口级限流：登录、支付、秒杀
- 热点参数限流：`product_id`
- 名单策略：黑名单直接拒绝，灰名单进入更严格限流
- 命中后常见业务码：`701` / `702`
