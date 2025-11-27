# SneakerFlash 后端接口文档（当前实现）

- 基础地址：`http://localhost:8000/api/v1`
- 鉴权：受保护接口需在 Header 携带 `Authorization: Bearer <access_token>`
- 统一返回：`{ "code": number, "msg": string, "data"?: any }`；`code=200` 表示业务成功。

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
  成功：`data=User`。
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

## 数据模型（核心字段）
- `User`：`id`, `user_name`, `avatar`, `created_at`, `updated_at`
- `Product`：`id`, `user_id`, `name`, `price`, `stock`, `start_time`, `image`, `created_at`, `updated_at`
- `Order`：`id`, `user_id`, `product_id`, `order_num`, `status`（0 未支付 /1 已支付 /2 失败）, `created_at`, `updated_at`
- `Payment`：`id`, `order_id`, `payment_id`, `amount_cents`, `status`（pending/paid/failed/refunded）, `notify_data?`, `created_at`, `updated_at`

## 风控与限流
- 开关：`config.yml` 中 `risk.enable`；按接口（登录/支付/秒杀）和热点参数（product_id）限流，命中返回 `code=701/702`。  
- 键粒度：默认按用户 ID / IP + 路径或参数建桶。修改频次可调整 `rate/burst/ttl` 配置。***
