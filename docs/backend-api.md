# SneakerFlash 后端接口文档（基于当前代码）

- 基础地址：`http://localhost:8000/api/v1`
- 鉴权：受保护接口需在 Header 中携带 `Authorization: Bearer <token>`
- 返回格式：除秒杀接口外，多数接口直接返回业务数据或 `{"error": "<msg>"}`；秒杀接口使用 `code` 字段表示业务态。

## 用户

- `POST /register`
  - Body：`{"user_name": string, "user_password": string}`
  - 成功：`200 {"message": "注册成功"}`
  - 失败：`400` 参数缺失；`500 {"error": "用户已存在"}` 等。

- `POST /login`
  - Body：同上。
  - 成功：`200 {"msg": "登录成功", "token": "<jwt>"}`（需保存并放入 Authorization）
  - 失败：`400` 参数缺失；`401 {"error": "用户不存在" | "密码错误"}`。

- `GET /profile` （鉴权）
  - 成功：`200 {"data": User}`
  - 失败：`401 {"error": "未登录"}`。

## 商品

- `GET /products`
  - Query：`page`（默认 1），`size`（默认 10）
  - 成功：`200 {"data": Product[], "total": number, "page": number}`

- `GET /product/:id`
  - 成功：`200 {"data": Product}`
  - 失败：`500 "找不到商品信息"` 等。

- `POST /products`（鉴权，发布商品）
  - Body：`{"name": string, "price": number, "stock": number, "start_time": string, "image": string}`
  - 成功：`200 {"msg": "商品发布成功", "data": Product}`
  - 失败：`500 {"error": "<数据库或校验错误>"}`。

## 秒杀

- `POST /seckill`（鉴权）
  - Body：`{"product_id": number}`
  - 成功：`200 {"code": 200, "msg": "抢购成功, 订单生成中", "data": {"order_num": string}}`
  - 失败：
    - `200 {"code": 500, "msg": "您已经抢购过该商品, 请勿重复下单" | "手慢无, 商品已经售罄" | "系统繁忙, 请稍后重试"}`（业务失败）
    - `401 {"error": "请先登录" | "token 无效"}`。

## 数据模型（对前端暴露的字段）

- `User`：`id`, `username`, `balance`, `created_at`, `updated_at`
- `Product`：`id`, `name`, `price`, `stock`, `start_time`, `image`, `created_at`, `updated_at`
- `Order`（由 worker 异步创建）：`id`, `user_id`, `product_id`, `order_num`, `status`（0: 未支付，1: 已支付，2: 失败），`created_at`, `updated_at`

## 行为说明

- JWT 失效或缺失：中间件直接返回 401。
- 秒杀成功会投递 Kafka 消息生成订单；前端仅需展示 `order_num`，无须轮询支付状态（当前代码未提供支付接口）。
