# SneakerFlash 测试设计方案

## 目标

- 建立明确的测试金字塔：单元测试 > 集成测试 > 端到端测试
- 先覆盖高风险链路，再补外围能力，避免一次性铺太大
- 测试结果可直接接入 CI，区分快速反馈与慢速验证
- 测试代码与业务代码一起演进，避免测试长期失效

## 当前现状

- 仓库当前没有提交任何 `*_test.go`
- 前端没有 `Vitest` / `Playwright` 配置
- 后端存在较多全局依赖：`db.DB`、`redis.RDB`、`kafka.Producer`、`config.Conf`
- `service` / `handler` 当前依赖具体实现而不是抽象接口，导致单测替身不容易注入
- `internal/integration/` 目录已存在但为空，适合作为集成测试入口

结论：当前最缺的不是“把测试全加上”，而是先定义测试分层、目录约定和依赖注入边界。

## 分层策略

### 1. 单元测试

目标：不依赖真实 MySQL / Redis / Kafka，强调快速、稳定、可并行。

覆盖范围：

- `internal/pkg/utils`
  - 密码哈希与校验
  - JWT 生成、解析、过期、类型校验
  - Snowflake 初始化与生成
- `internal/pkg/app`
  - 统一响应结构序列化
- `internal/middlerware`
  - JWT 中间件
  - 本地限流 / 接口限流 / 参数限流
  - metrics 中间件计数逻辑
- `internal/service`
  - `user.go`：注册、登录、刷新 token、更新资料
  - `order.go`：轮询状态映射、支付状态校验、优惠券应用分支
  - `coupon.go`：可用性判定、过期与核销边界
  - `vip.go`：成长值、权益、等级计算
  - `health.go`：健康检查 / 就绪检查状态聚合
- `internal/handler`
  - 参数绑定失败
  - 业务错误码映射
  - 成功响应结构

设计要求：

- `service` 层逐步改为依赖窄接口，而不是直接依赖具体 repo
- 时间、ID 生成、消息发送、缓存访问通过注入函数或接口隔离
- 单测统一使用表驱动，命名模式：`Test<Service>_<Method>_<Scenario>`

### 2. 集成测试

目标：验证多个模块协同时的真实行为，但范围仍然可控。

覆盖范围：

- `repository + GORM + MySQL`
  - 用户唯一键
  - 订单查询 / 支付单查询
  - 批量插入与事务回滚
- `Redis + Lua`
  - 秒杀库存扣减
  - 防重复购买
  - pending 订单缓存读写
  - 限流器 token bucket 行为
- `Kafka + Outbox + Worker`
  - Outbox 写入成功后异步发送
  - Worker 消费后创建订单 / 支付单
  - 消费幂等
  - 失败消息重试 / 死信
- `Gin Router + Middleware + Handler + Service`
  - 登录、刷新、鉴权、个人资料
  - 秒杀 -> 轮询 -> 支付回调 主链路
  - `health` / `ready` 可用性探针

建议位置：

- 后端集成测试统一放在 `internal/integration/`
- 按主题拆分：
  - `internal/integration/auth_test.go`
  - `internal/integration/seckill_flow_test.go`
  - `internal/integration/payment_flow_test.go`
  - `internal/integration/risk_limit_test.go`

建议运行方式：

- 使用 Docker Compose 或 `testcontainers-go` 拉起 MySQL / Redis / Kafka
- 使用单独测试库、测试 topic、测试 Redis key 前缀
- 集成测试默认通过 build tag 区分：`go test -tags=integration ./internal/integration/...`

### 3. 端到端测试

目标：从用户视角验证前后端联调结果，确保页面、路由、请求、状态更新一致。

覆盖范围：

- 认证链路
  - 注册
  - 登录
  - token 过期后自动刷新
  - 未登录访问受保护页面跳转
- 商品链路
  - 首页列表展示
  - 商品详情加载
  - 发布商品
  - 我的商品编辑 / 删除
- 秒杀链路
  - 秒杀按钮状态
  - 提交后进入 pending
  - 轮询 ready
  - 订单详情页展示
- 支付链路
  - 支付回调后订单状态变化
  - 优惠券使用后金额变化
- 用户链路
  - 头像上传
  - 个人资料更新
  - VIP 页面权益显示

前端测试再分两层：

- 组件 / store / router 测试：`Vitest + Vue Test Utils`
- 浏览器级别 E2E：`Playwright`

建议目录：

- `frontend/src/**/*.test.ts`
- `frontend/tests/e2e/**/*.spec.ts`

## 推荐的测试目录结构

```text
internal/
├── integration/
│   ├── auth_test.go
│   ├── seckill_flow_test.go
│   ├── payment_flow_test.go
│   └── helpers/
│       ├── app.go
│       ├── fixtures.go
│       └── env.go
├── middlerware/
│   └── *_test.go
├── pkg/
│   └── **/*_test.go
├── repository/
│   └── *_test.go
└── service/
    └── *_test.go

frontend/
├── src/
│   ├── lib/api.test.ts
│   ├── router/index.test.ts
│   ├── stores/*.test.ts
│   └── views/**/*.test.ts
└── tests/
    ├── e2e/
    │   ├── auth.spec.ts
    │   ├── seckill.spec.ts
    │   └── order-payment.spec.ts
    └── fixtures/
```

## 各层重点用例

### P0：必须先补

这些是系统最值钱、最容易出事故的链路。

- `UserService`
  - 重复注册
  - 密码错误登录
  - refresh token 非 refresh 类型
  - 更新资料时用户名冲突
- `JWTauth`
  - 缺少 token
  - token 格式错误
  - token 过期
  - access / refresh 类型混用
- `SeckillService`
  - 活动未开始
  - 活动已结束
  - 重复抢购
  - 库存售罄
  - 写 Outbox 失败时 Redis 回滚
- `WorkerService`
  - 消息解析失败
  - 幂等消费
  - 批量扣库存部分失败
  - 支付单缺失时自动补全
- `OrderService`
  - `PollOrder` 的 `pending/ready/failed`
  - 支付回调幂等
  - 非本人订单不可见
  - 已支付订单不可再改优惠券
- 集成链路
  - 秒杀成功后可轮询到 ready
  - 支付成功后订单状态变更
  - 限流命中返回 429/7xx

### P1：第二批补充

- `CouponService` 优惠券状态机
- `VIPService` 成长值与会员权益
- `ProductService` 发布、修改、删除权限
- `health` / `ready` 探针
- `metrics` 指标暴露
- 前端 `api.ts` 自动刷新 token 队列
- `userStore` / `productStore` 状态切换
- router 鉴权跳转

### P2：后续增强

- 异常注入测试：Kafka 不可用、Redis 超时、数据库死锁
- Outbox 补偿任务
- DLQ 重放
- 并发秒杀稳定性测试
- 浏览器端多标签页 token 刷新竞争

## 当前代码为可测试性需要做的最小改造

这是落地测试前最关键的一步。

### 1. 后端服务层改为依赖接口

当前问题：

- `UserService` 直接依赖 `*repository.UserRepo`
- `OrderService` 直接依赖多个具体 repo
- `SeckillService` 直接读写全局 `redis.RDB` / `kafka.Send`

建议改造：

- 在 `service` 层定义窄接口，例如：
  - `UserReader`
  - `UserWriter`
  - `OrderStore`
  - `PaymentStore`
  - `StockCache`
  - `MessageBus`
- 构造函数依赖接口，不依赖具体 repo

收益：

- 单测可以直接塞内存 stub / fake
- 不必每个服务单测都起真实数据库

### 2. 隔离全局状态

优先去掉这些测试阻塞点：

- `config.Conf`
- `db.DB`
- `redis.RDB`
- `kafka.Producer`
- `time.Now()`
- `utils.GenSnowflakeID()`

建议方式：

- 将配置封装成显式依赖
- 将当前时间与 ID 生成器抽象成函数注入
- Redis / Kafka 操作封装到网关接口

### 3. 为 handler 提供 mockable service 接口

当前 handler 也直接依赖具体 service。

建议：

- 在 `handler` 包内定义最小接口
- `NewUserHandler` / `NewOrderHandler` 接受接口而不是具体 struct

这样可以用 `httptest` 精准验证：

- 参数校验
- 错误码映射
- JSON 输出格式

## 测试数据与环境约定

### 数据工厂

统一提供 builder / fixture：

- `NewTestUser()`
- `NewTestProduct()`
- `NewTestOrder()`
- `NewPendingOrderCache()`

避免每个测试手写大量样板数据。

### 数据隔离

- MySQL：每个测试用例使用事务回滚，或每个测试文件使用独立 schema
- Redis：统一前缀 `test:<suite>:<case>:*`
- Kafka：topic 增加测试后缀，consumer group 使用唯一值

### 稳定性要求

- 禁止依赖真实线上配置
- 禁止复用开发环境数据库
- 测试必须可重复执行
- E2E 必须支持 headless 跑在 CI

## 前端测试方案

### 组件 / Store / Router

推荐工具：

- `Vitest`
- `@vue/test-utils`
- `@pinia/testing`
- `msw` 或 `axios-mock-adapter`

重点覆盖：

- `frontend/src/lib/api.ts`
  - `code != 200` 抛错
  - 401 自动 refresh
  - refresh 并发队列只触发一次
  - refresh 失败后清 token 并跳转登录
- `frontend/src/router/index.ts`
  - `requiresAuth` 跳转登录
  - 带 `redirect` 参数
- `frontend/src/stores/userStore.ts`
  - 登录成功持久化 token
  - `fetchProfile` 失败后清理状态
  - `logout` 清空本地状态
- `frontend/src/stores/productStore.ts`
  - 列表刷新
  - detail cache 命中
  - 删除后刷新我的商品与首页列表

### Playwright E2E

建议先做 3 条主用例：

- `auth.spec.ts`
  - 注册 -> 登录 -> 跳首页
- `seckill.spec.ts`
  - 进入商品详情 -> 发起秒杀 -> 轮询 ready -> 跳订单页
- `order-payment.spec.ts`
  - 查看订单 -> 模拟支付回调 -> 状态变为已支付

## 命令与 CI 设计

### 本地命令

建议补充这些命令：

```bash
make test-unit
make test-integration
make test-e2e
make test-all
```

对应含义：

- `test-unit`：Go 单测 + 前端 Vitest
- `test-integration`：后端集成测试
- `test-e2e`：Playwright 端到端测试
- `test-all`：全量执行

### CI 分层

#### PR 必跑

- `golangci-lint run ./...`
- `go test` 快速单测
- `pnpm lint`
- `pnpm build`
- `pnpm vitest run`

#### 合并前或主干必跑

- 后端集成测试
- Swagger 与接口文档一致性检查

#### 夜间任务

- Playwright 全量 E2E
- k6 smoke 压测

## 覆盖率目标

不要一开始追求全仓库高覆盖率，先追求核心路径可信。

- 第一阶段：后端核心 service 覆盖率 >= 60%
- 第二阶段：`service + middlerware + handler` >= 70%
- 第三阶段：前端 `lib/router/store` 关键模块 >= 70%
- 端到端只看关键业务流是否覆盖，不以行覆盖率为目标

## 推荐的落地顺序

### 阶段 1：一周内完成

- 建后端单测基础设施
- 先补 `user`、`order`、`jwt`、`health`
- 建前端 `Vitest` 基础设施
- 先补 `api.ts`、`userStore`、router

### 阶段 2：两周内完成

- 补 `seckill`、`worker`、`coupon` 核心集成测试
- 拉起 MySQL / Redis / Kafka 的测试环境
- 接入 CI 分层执行

### 阶段 3：后续持续演进

- 引入 Playwright
- 补齐商品发布、秒杀、支付 E2E
- 增加异常场景与补偿链路验证

## 我对这个仓库的具体建议

如果只做一件最有价值的事，我建议先做下面这一组：

1. 后端先补 `UserService`、`OrderService.PollOrder`、`JWTauth`
2. 再补一条 `seckill -> worker -> poll -> payment callback` 集成测试
3. 前端先补 `api.ts` 的 401 refresh 队列测试

原因很简单：

- 这三部分覆盖了登录、鉴权、异步建单、支付结果回写
- 它们正好是当前系统最核心、最容易出现回归的路径
- 成本可控，不需要一开始就重构整个仓库
