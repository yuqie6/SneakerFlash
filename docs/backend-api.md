# SneakerFlash 后端接口文档（当前实现）

- 基础地址：业务接口 `http://localhost:8000/api/v1`
- 探针接口：`http://localhost:8000/health`、`http://localhost:8000/ready`
- 鉴权：受保护接口需在 Header 携带 `Authorization: Bearer <access_token>`
- 统一返回：`{ "code": number, "msg": string, "data"?: any }`；`code=200` 表示业务成功。

## 系统
- `GET /health`（根路径，非 `/api/v1`）
  快速存活检查，仅确认 HTTP 服务进程可响应。  
  成功：`code=200`，`data={"status":"ok","service":"SneakerFlash","timestamp":RFC3339}`。
- `GET /ready`（根路径，非 `/api/v1`）
  就绪检查，校验 MySQL、Redis 和 Kafka Producer 初始化状态。  
  成功：`code=200`，`data={"status":"ready","checks":{"database":{"status":"up"},"redis":{"status":"up"},"kafka":{"status":"up"}}}`。  
  未就绪：HTTP `503`，`msg="服务未就绪"`，并在 `data.checks` 返回失败组件明细。

## 认证
- `POST /register`  
  Body：`{ "user_name": string, "user_password": string }`  
  成功：`code=200`，`data={"message":"注册成功"}`；已存在返回 `code=10001`。
- `POST /login`  
  Body 同上。  
  成功：`code=200`，`data={ "access_token", "refresh_token", "expires_in" }`。  
  用户不存在/密码错误返回 `401` + 对应业务码。
- `POST /refresh`  
  Body：`{ "refresh_token": string }`  
  成功：`code=200`，`data={ "access_token", "expires_in" }`。

## 用户
- `GET /profile`（鉴权）  
  成功：`data=User`，当前会返回 `total_spent_cents`、`growth_level`、`role`。
- `PUT /profile`（鉴权）  
  Body：可选 `user_name`, `avatar`；至少传一项。  
  成功：`data=User`；重名返回 `code=10001`。

## 上传
- `POST /upload`（鉴权）`multipart/form-data`，字段 `file`；成功返回 `data={ "url": "<path>" }`。

## 商品
- `GET /products?page&size`  
  成功：`data={ items: Product[], total, page }`。
- `GET /product/:id`  
  成功：`data=Product`；不存在返回 `404` + `code=20001`。
- `POST /products`（鉴权）  
  Body：`name` `price` `stock` `start_time`(未来时间) `image?`；成功返回 `data=Product`。
- `PUT /products/:id`（鉴权，仅发布者）  
  Body：任意字段可选（同上）；成功返回 `data={id}`。
- `DELETE /products/:id`（鉴权，仅发布者）  
  成功：`data={id}`。
- `GET /products/mine`（鉴权）  
  Query：`page` `size`；成功：`data={ items, total, page, size }`。

## 秒杀
- `POST /seckill`（鉴权）  
  Body：`{ "product_id": number }`  
  成功：`code=200`，`data={"order_num": string, "payment_id": string, "status": "pending"|"ready"}`。  
  业务失败：`code=30001`（售罄）、`30002`（重复下单）、`30003`（限流）。

## 订单 & 支付
- `GET /orders`（鉴权）  
  Query：`page` `size`，可选 `status`（0=未支付，1=已支付，2=失败）。  
  成功：`data={ items: Order[], total, page, size }`。
- `GET /orders/:id`（鉴权，需本人）  
  成功：`data={ order: Order, payment?: Payment }`。
- `GET /orders/poll/:order_num`（鉴权）  
  轮询异步秒杀订单状态；pending 返回 `{status,payment_id}`，ready 返回 `{status,order}`。
- `POST /payment/callback`  
  Body：`{ "payment_id": string, "status": "paid"|"failed"|"refunded", "notify_data"?: string }`  
  成功：`data={ order, payment }`；未找到支付单返回 `404`。

## 管理后台
- 鉴权要求：所有 `/admin/*` 接口都需要管理员 `access_token`；普通用户会收到 HTTP `403` + `msg="需要管理员权限"`。
- `GET /admin/stats`  
  成功：`data={ total_users, total_orders, total_revenue_cents, total_products, pending_orders }`。
- `GET /admin/users?page=1&page_size=20`  
  成功：`data={ list: User[], total, page, page_size }`，返回全站用户。
- `GET /admin/orders?page=1&page_size=20&status=0|1|2`  
  成功：`data={ list: Order[], total, page, page_size }`，返回全站订单。
- `GET /admin/products?page=1&page_size=20`  
  成功：`data={ list: Product[], total, page, page_size }`，返回全站商品。
- `GET /admin/coupons?page=1&page_size=20`  
  成功：`data={ list: Coupon[], total, page, page_size }`。
- `POST /admin/coupons`  
  Body：`{ type, title, description, amount_cents, discount_rate, min_spend_cents, valid_from, valid_to, purchasable, price_cents, status }`。  
  `type` 支持 `full_cut|discount`，`status` 支持 `active|inactive`；时间支持 `RFC3339`、`YYYY-MM-DD HH:mm[:ss]`、`YYYY-MM-DDTHH:mm[:ss]`。  
  成功：`data=Coupon`。
- `PUT /admin/coupons/:id`  
  Body 同上，支持部分字段更新。成功：`data=Coupon`。
- `DELETE /admin/coupons/:id`  
  成功：`data={ "message": "ok" }`。
- `GET /admin/risk/blacklist` / `GET /admin/risk/graylist`  
  成功：`data={ ip: string[], user: string[] }`。
- `POST /admin/risk/blacklist` / `POST /admin/risk/graylist`  
  Body：`{ "type": "ip"|"user", "value": string }`；成功：`data={ "message": "ok" }`。
- `DELETE /admin/risk/blacklist` / `DELETE /admin/risk/graylist`  
  Body 同上；成功：`data={ "message": "ok" }`。

## 数据模型（核心字段）
- `User`：`id`, `username`, `balance`, `avatar`, `total_spent_cents`, `growth_level`, `role`, `created_at`, `updated_at`
- `Product`：`id`, `user_id`, `name`, `price`, `stock`, `start_time`, `image`, `created_at`, `updated_at`
- `Order`：`id`, `user_id`, `product_id`, `order_num`, `status`（0 未支付 /1 已支付 /2 失败）, `created_at`, `updated_at`
- `Payment`：`id`, `order_id`, `payment_id`, `amount_cents`, `status`（pending/paid/failed/refunded）, `notify_data?`, `created_at`, `updated_at`
- `Coupon`：`id`, `type`, `title`, `description`, `amount_cents`, `discount_rate`, `min_spend_cents`, `valid_from`, `valid_to`, `purchasable`, `price_cents`, `status`

## 风控与限流
- 开关：应用配置文件中的 `risk.enable`；按接口（登录/支付/秒杀）和热点参数（product_id）限流，命中返回 `code=701/702`。  
- 键粒度：默认按用户 ID / IP + 路径或参数建桶。修改频次可调整 `rate/burst/ttl` 配置。
