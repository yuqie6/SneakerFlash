# 阶段路线图

- **P0**：把现有接口稳定下来，补齐安全基础。
- **P1**：做完订单支付、限流风控、监控、VIP 和优惠券。
- **P2**：高可用、未支付自动取消、管理后台。

## P0：接口稳定与基础安全

1) 接口规范
   - 统一响应格式 `{ code, msg, data }`
   - 定义错误码：401 未登录、429 限流、5xx 系统错误、7xx 风控拦截
   - CORS 配置

2) 鉴权
   - JWT 短 token + refresh token
   - 中间件统一注入 userID/username
   - 密码策略（可选）

3) 商品与库存
   - 发布校验：开始时间、库存、价格
   - 预热失败时回滚

4) 秒杀
   - Lua 脚本防重复购买（已有）
   - Kafka 发送失败重试，或写入 Outbox 由定时任务补发
   - Worker 建单事务失败时回滚 Redis 库存

5) 配置
   - dev/stage/prod 分环境配置
   - 敏感信息通过环境变量注入

## P1：订单支付、限流、监控、VIP 与优惠券

1) 订单与支付
   - `GET /orders`（分页、按状态筛选）、`GET /orders/:id`
   - 生成支付单，支付回调写入数据库（待支付 → 已支付/失败）
   - 用 order_num + 用户 ID 防重复

2) 限流与风控
   - 接口级限流：登录、秒杀、支付各自独立的令牌桶
   - 热点参数限流：按 product_id 限制 QPS，防单个商品打爆
   - 用户/IP 黑名单，异常请求返回 429

3) 监控
   - 指标：QPS、成功率、P95/P99 延迟、库存命中率、Kafka 消费延迟
   - 日志：slog + lumberjack，结构化输出，带 trace-id 和 uid
   - 告警：Kafka 消费积压、Redis 异常、DB 慢查询、错误率超阈值（Prometheus + Grafana）

4) 性能
   - Redis 连接池和超时配置
   - MySQL 索引（product_id、user_id、order_num）
   - 队列消费失败重试 + 死信主题
   - 缓存过期时间加随机偏移，防止大量 key 同时过期
   - 对不存在的数据缓存空值，防止反复查数据库

5) VIP 与优惠券
   - 成长等级按累计实付金额计算，永久生效
   - 可购买付费 VIP，有效等级取成长等级和付费 VIP 中较高的
   - 优惠券支持满减和折扣，按月发放，VIP 等级越高配额越多
   - 普通用户也可以购买优惠券
   - 优惠券当月有效，每个订单只能用一张
   - 成长等级分层（按累计实付元）：L1 0–999 / L2 1,000–4,999 / L3 5,000–19,999 / L4 20,000+

## P2：高可用与运营

1) 熔断降级
   - DB/Redis/Kafka 响应慢时自动熔断，返回"系统繁忙"

2) 优雅停机
   - API 和 Worker 收到信号后停止接受新请求，等当前请求处理完再退出
   - Kafka consumer 关闭前确认 offset

3) 未支付订单自动取消
   - 15 分钟未支付的订单自动取消，回滚库存
   - 可用 Redis ZSet 或 RabbitMQ 延时队列实现

4) 消息补偿
   - 秒杀成功后先写本地消息表（状态=待发送），定时任务扫描重试 Kafka

5) 实时推送
   - WebSocket/SSE 推送库存变化和订单状态

6) 管理后台
   - 商品、库存、订单、风控看板
   - 权限分级（RBAC）

7) 数据安全
   - 审计日志（发布、库存变更、订单状态变更）
   - 少存个人敏感信息，密钥统一管理

## 接口清单

现有：
- `POST /register`、`POST /login`、`GET /profile`
- `GET /products`、`GET /product/:id`、`POST /products`（需登录）
- `POST /seckill`（需登录）

P1 新增：
- `GET /orders`、`GET /orders/:id`
- `POST /orders/:id/pay` 或 `/pay`
- `POST /payment/callback`（需签名校验）
- 健康检查（可选）

## 数据模型

- User：`id, username, password_hash, balance, created_at, updated_at`
- Product：`id, name, price, stock, start_time, image, created_at, updated_at`
- Order：`id, user_id, product_id, order_num, status (0/1/2), created_at, updated_at`
- 可扩展：支付单、风控记录、库存日志、本地消息表

## 技术备忘

- Redis Lua 原子扣库存 + 用户去重
- Kafka：消费失败重试 + 死信主题 + 消费组监控；发送失败走本地消息表补偿
- MySQL：短事务，索引优化；订单创建和库存扣减在同一个事务里
- 配置：Viper + 环境变量覆盖；敏感信息不入代码
- HTTP：CORS、gzip、合理超时；日志带 trace-id

## 建议交付顺序

1. **P0**：统一响应格式和错误码 → 鉴权 → 发布校验 → CORS 和分环境配置
2. **P1**：订单支付 → 限流风控 → 监控（slog + Prometheus + Grafana）→ 消费重试和死信
3. **P2**：熔断降级 → 未支付自动取消 → 消息补偿 → 优雅停机 → 实时推送 → 管理后台
