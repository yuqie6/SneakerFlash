# 阶段路线图（审计于 2026-03-18）

## 结论
- **P0：已完成**。接口规范、JWT + refresh token、商品发布校验、Redis 预热回滚、分环境配置均已落地。
- **P1：已基本完成**。订单查询、支付回调、接口级/热点参数限流、黑灰名单、Prometheus 指标、结构化日志、VIP/优惠券、Outbox + DLQ 已具备可用实现。
- **P2：已推进到可用基线**。轻量熔断、优雅停机、未支付自动取消、SSE 实时推送、审计日志、资源级 RBAC 已落地；告警模板与更深的高可用治理仍待补强。

## P0：接口稳定与基础安全
### 已完成
1. 接口规范
   - 统一响应格式 `{ code, msg, data }`
   - 已定义未登录、限流、业务错误和风控错误码
   - CORS 已在 HTTP Server 中统一配置

2. 鉴权
   - 已实现 access token + refresh token
   - JWT 中间件统一注入 `userID`、`username`、`role`
   - 管理员接口基于 `AdminAuth + AdminResourceAuth` 做角色和资源校验

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

2. 优雅停机
   - API 已接入 `SIGINT` / `SIGTERM` 优雅退出，按 HTTP -> Kafka -> Redis -> DB 顺序收口
   - Worker 已接入信号处理，会先停 cron、再退出 consumer、最后关闭底层连接

3. 自动取消与实时同步
   - Worker 已实现 15 分钟未支付订单自动取消，并回补库存、释放用户标记和优惠券
   - 已新增 SSE 订单状态 / 商品库存推送；前端保留轮询兜底

4. 数据安全与治理
   - 后台已支持资源级 RBAC：`admin`、`ops_admin`、`risk_admin`、`coupon_admin`、`audit_admin`
   - 已新增审计日志模型与 `/admin/audit` 查询接口，优惠券和风控名单操作会落审计

5. 轻量熔断降级
   - 已对 Redis 秒杀入口与 Kafka Producer 即时发送补入轻量熔断，指标会暴露状态切换与拒绝次数

### 仍需补强
- 轻量熔断目前主要覆盖 Redis / Kafka Producer，尚未扩展到更细粒度的 DB / consumer 依赖点
- 监控面板样例、告警阈值和标准化告警策略仍主要停留在运维基线层面
- SSE 目前是单实例内存广播模型，适合作为体验增强，不适合作为多实例事实通道

## 当前接口清单（与实现一致）
- 认证：`POST /register`、`POST /login`、`POST /refresh`
- 用户：`GET /profile`、`PUT /profile`、`POST /upload`
- 推送：`GET /stream/orders/:id`、`GET /stream/products/:id`
- 商品：`GET /products`、`GET /product/:id`、`POST /products`、`PUT /products/:id`、`DELETE /products/:id`、`GET /products/mine`
- 秒杀：`POST /seckill`
- 订单：`GET /orders`、`GET /orders/:id`、`GET /orders/poll/:order_num`、`POST /orders/:id/apply-coupon`
- 支付：`POST /payment/callback`
- VIP：`GET /vip/profile`、`POST /vip/purchase`
- 优惠券：`GET /coupons/mine`、`POST /coupons/purchase`
- 管理后台：`GET /admin/stats`、`GET /admin/users`、`GET /admin/orders`、`GET /admin/products`、`GET /admin/coupons`、`POST /admin/coupons`、`PUT /admin/coupons/:id`、`DELETE /admin/coupons/:id`、`GET/POST/DELETE /admin/risk/blacklist|graylist`、`GET /admin/audit`

## 下一阶段建议
1. 先补监控面板样例、告警阈值和运维模板，让当前 P2 基线真正可观测。
2. 再决定是否把熔断和 SSE 从单实例能力扩展到多实例场景。
3. 若继续演进后台治理，再从“资源级权限”细化到“动作级权限”。
