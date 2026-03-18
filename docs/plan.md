# 阶段路线图（审计于 2026-03-18）

## 结论
- **P0：已完成**。接口规范、JWT + refresh token、商品发布校验、Redis 预热回滚、分环境配置均已落地。
- **P1：已基本完成**。订单查询、支付回调、接口级/热点参数限流、黑灰名单、Prometheus 指标、结构化日志、VIP/优惠券、Outbox + DLQ 已具备可用实现。
- **P2：部分完成**。管理后台和消息补偿已落地；熔断降级、优雅停机、未支付自动取消、实时推送、审计日志、细粒度 RBAC 仍未完成。

## P0：接口稳定与基础安全
### 已完成
1. 接口规范
   - 统一响应格式 `{ code, msg, data }`
   - 已定义未登录、限流、业务错误和风控错误码
   - CORS 已在 HTTP Server 中统一配置

2. 鉴权
   - 已实现 access token + refresh token
   - JWT 中间件统一注入 `userID`、`username`、`role`
   - 管理员接口基于 `AdminAuth` 做角色校验

3. 商品与库存
   - 发布时校验开始时间、结束时间、价格和库存
   - 商品创建后会把库存预热到 Redis，预热失败会回滚数据库记录
   - 商品详情优先读取缓存，并以 Redis 实时库存为准

4. 秒杀
   - Redis Lua 已实现判重 + 扣库存原子操作
   - 秒杀入口已写 Outbox，异步发送 Kafka
   - Worker 建单失败时会回滚 Redis 库存与用户标记

5. 配置
   - 已统一为 `SNEAKERFLASH_CONFIG -> config.<env>.local.yml`
   - 已提供 dev/prod example 与 `.env.*.example`
   - Viper 已启用环境变量覆盖

## P1：订单支付、限流、监控、VIP 与优惠券
### 已完成
1. 订单与支付
   - `GET /orders`、`GET /orders/:id`、`GET /orders/poll/:order_num` 已实现
   - 支付回调 `POST /payment/callback` 已实现幂等状态推进
   - `order_num`、`payment_id`、订单状态条件更新已用于防重复处理
   - 支付前支持 `POST /orders/:id/apply-coupon`

2. 限流与风控
   - 登录、支付、秒杀已支持接口级限流
   - `product_id` 已支持热点参数限流
   - 黑名单 / 灰名单已支持按用户和 IP 管控，并具备管理端接口

3. 监控
   - 已暴露 `/metrics`
   - 已接入 `slog + lumberjack`
   - Dev/Prod Compose 已包含 Prometheus + Grafana 基线
   - 健康检查已提供 `/health`、`/ready`

4. 性能与可靠性
   - Redis 连接池、超时、Kafka consumer offset 策略已可配置
   - Outbox 补偿、消费者重试、DLQ 已实现
   - 商品缓存已包含随机 TTL 与空值缓存策略
   - MySQL 索引与慢查询阈值已有配置约束

5. VIP 与优惠券
   - 成长等级基于累计实付金额
   - 付费 VIP 已实现，并按成长等级/付费等级取较高值生效
   - 优惠券已支持满减、折扣、购买、月度发放、订单核销与释放
   - 管理后台已支持优惠券模板 CRUD

### 仍需补强
- 监控面板样例、告警阈值和标准化告警策略仍主要停留在运维基线层面，尚未形成成套文档与配置模板。

## P2：高可用与运营
### 已完成
1. 消息补偿
   - 秒杀入口先写 Outbox，再异步发 Kafka
   - Worker 侧与补偿任务均已支持失败重试和 DLQ

2. 管理后台
   - 已具备统计、用户、订单、商品、优惠券、风控名单管理接口
   - 前端管理页也已接入对应路由与页面

### 未完成
1. 熔断降级
   - 未发现 DB / Redis / Kafka 级别的熔断器实现

2. 优雅停机
   - `cmd/api` 仍直接 `Run()`，`cmd/worker` 仍直接启动 consumer，未接入信号处理与优雅退出流程

3. 未支付订单自动取消
   - 未发现 15 分钟自动取消订单并回滚库存的任务或延迟队列实现

4. 实时推送
   - 未发现 WebSocket / SSE 库存与订单状态推送实现

5. 数据安全与治理
   - 目前只有 `role=admin` 的单级管理员模型，尚未形成细粒度 RBAC
   - 未发现订单/库存变更等审计日志模型与查询链路

## 当前接口清单（与实现一致）
- 认证：`POST /register`、`POST /login`、`POST /refresh`
- 用户：`GET /profile`、`PUT /profile`、`POST /upload`
- 商品：`GET /products`、`GET /product/:id`、`POST /products`、`PUT /products/:id`、`DELETE /products/:id`、`GET /products/mine`
- 秒杀：`POST /seckill`
- 订单：`GET /orders`、`GET /orders/:id`、`GET /orders/poll/:order_num`、`POST /orders/:id/apply-coupon`
- 支付：`POST /payment/callback`
- VIP：`GET /vip/profile`、`POST /vip/purchase`
- 优惠券：`GET /coupons/mine`、`POST /coupons/purchase`
- 管理后台：`GET /admin/stats`、`GET /admin/users`、`GET /admin/orders`、`GET /admin/products`、`GET /admin/coupons`、`POST /admin/coupons`、`PUT /admin/coupons/:id`、`DELETE /admin/coupons/:id`、`GET/POST/DELETE /admin/risk/blacklist|graylist`

## 下一阶段建议
1. 先补 P2 中最影响稳定性的三项：优雅停机、未支付订单自动取消、熔断降级。
2. 再完善运营侧能力：告警模板、审计日志、细粒度 RBAC。
3. 实时推送建议最后做，避免在核心一致性链路未完全收口前引入额外复杂度。
