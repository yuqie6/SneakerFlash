# 阶段路线图
- **P0 必须**：标准化接口与安全，稳定现有秒杀/发布链路。
- **P1 交易闭环**：订单/支付、风控/限流、幂等一致性、可观测。
- **P2 高可用/运营**：实时推送、延时/补偿、后台与治理。

## P0：接口稳定与基础安全
1) 接口契约与错误码
   - 统一响应：`{ code, msg, data }`；HTTP 语义保留。
   - 错误码枚举：401 未登录；429 频控；5xx 业务/系统；7xx 风控。
   - CORS/OPTIONS 正式化。
2) 鉴权与安全
   - JWT 失效策略（短 token + 明确过期）；必要时引入 refresh。
   - 身份中间件统一注入 userID/username；密码策略与封禁（可选）。
3) 商品与库存
   - 发布校验：开始时间晚于当前、库存/价格合法、防重复提交。
   - 预热失败回滚/告警。
4) 秒杀链路
   - Lua 幂等（重复购买校验已做）。
   - Kafka 发送失败重试/补偿或死信。
   - Worker 日志与告警；订单事务失败回滚库存。
5) 配置与环境
   - 分环境配置（dev/stage/prod），敏感信息 env 注入。

## P1：交易闭环与风控
1) 订单/支付接口（新增）
   - `GET /orders`（分页/状态），`GET /orders/:id`
   - 支付发起：生成支付单/签名，返回前端；支付回调落库（待支付 -> 已支付/失败）。
   - 幂等键：order_num + 用户，防重复落库/重复回调。
2) 风控与限流
   - 接口级限流（登录/秒杀/支付令牌桶）。
   - 热点参数限流：按 product_id QPS（如 1000/s），防爆款打挂。
   - 设备/账号频次监控，异常返回 429/风控码；IP 黑/灰名单可插拔。
3) 可观测性
   - 指标：QPS、成功率、p95/p99、库存命中率、Kafka 投递/消费延迟。
   - 日志：zap + lumberjack，结构化日志，trace-id/uid 贯穿。
   - 告警：Kafka lag、Redis 失败、DB 慢查、错误率阈值（Prometheus + Alertmanager/Grafana）。
4) 性能与可靠性
   - Redis 连接池/超时；MySQL 索引（product_id/user_id/order_num）；慢查询监控。
   - 队列重试/死信；缓存随机过期防雪崩；null 缓存防穿透。

## P2：高可用、补偿与运营
1) 熔断/降级
   - 引入 Sentinel/熔断组件：DB/Redis/Kafka 慢时熔断，降级返回“系统繁忙”。
   - 热点参数限流已上，结合降级策略。
2) 优雅停机
   - API 与 Worker 捕获信号，停止接收新流量，等待 in-flight 完成后退出；Kafka consumer Close 确认 offset。
3) 延时队列与补偿
   - 延时取消未支付订单（15 分钟），回滚库存：Redis ZSet 或 RabbitMQ DLX（Kafka 需额外实现）。
   - 本地消息表：Seckill 成功先写“待发送”消息，定时扫描重试 Kafka，保证最终一致性。
4) 实时推送
   - WebSocket/SSE 推库存/订单状态，断线重连与订阅管理。
5) 运营后台与治理
   - 管理后台（商品/库存/订单/风控看板），RBAC 分权。
   - A/B 灰度：限流/风控策略、交互实验。
6) 数据与合规
   - 审计日志（发布、库存调整、订单状态变更）。
   - PII 最小化存储，密钥管理。

## 接口清单
- 现有：`POST /register`，`POST /login`，`GET /profile`，`GET /products`，`GET /product/:id`，`POST /products`（鉴权），`POST /seckill`（鉴权）。
- 拟新增（订单/支付/风控）
  - `GET /orders`，`GET /orders/:id`
  - `POST /orders/:id/pay` 或 `/pay`（按渠道设计）
  - 支付回调 `POST /payment/callback`（签名校验）
  - （可选）风控/限流状态查询，健康检查。

## 数据与模型
- User：`id, username, password_hash, balance, created_at, updated_at`
- Product：`id, name, price, stock, start_time, image, created_at, updated_at`
- Order：`id, user_id, product_id, order_num, status (0/1/2), created_at, updated_at`
- 可扩展：payment_intent、风控记录、库存日志、消息表（待发送/已发送）。

## 技术要点备忘
- Redis Lua 原子扣减 + 用户集合去重；脚本幂等 & 错误码。
+- Kafka：重试、死信主题、消费组监控；发送失败回滚或本地消息表补偿。
- MySQL：短事务，索引优化；订单创建与扣库存同事务。
- 配置：viper + env；敏感信息不入库。
- HTTP：CORS/OPTIONS、gzip、合理超时；日志/trace 注入。

## 交付顺序（建议）
1) **P0**：标准化响应/错误码，鉴权策略，发布校验与告警，CORS/配置分环境。
2) **P1**：订单/支付接口与幂等；风控与热点限流；可观测（zap + Prometheus + Grafana + Jaeger）；队列重试/死信。
3) **P2**：熔断/降级（Sentinel）、延时队列取消未支付、消息表补偿、优雅停机、实时推送、后台与灰度治理。
