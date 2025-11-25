# Repository Guidelines

## 项目概览
- 后端：Go（Gin + GORM + Redis + Kafka + Viper），入口 `cmd/api`（HTTP）与 `cmd/worker`（Kafka 消费），业务与基础设施在 `internal/` 分层。
- 前端：Vue 3 + TypeScript + Vite + Tailwind，Pinia + Vue Router，动效 Lenis/Motion-v，UI 组件位于 `frontend/src/components/ui` 与 `components/motion`，入口 `frontend/src/main.ts`。
- 文档：接口 `docs/backend-api.md`、前端方案 `docs/frontend-plan.md`、阶段路线 `docs/plan.md`、压测指南 `docs/perf.md`，Swagger 产物在 `docs/swagger.*`。
- 压测/工具：`perf/k6-seckill.js`、`perf/export_tokens.go`；静态上传目录 `uploads/`（git 忽略）。

## 运行与配置
- Go >=1.22（go.mod 声明 1.25.1）；首次运行 `go mod tidy`。
- 配置：默认读取仓库根的 `config.yml`（已被 .gitignore，需本地自建），或设置环境变量 `SNEAKERFLASH_CONFIG` 指向自定义 yml；核心字段示例：
  ```yaml
  server: { port: ":8000", machineid: 1, upload_dir: "uploads" }
  data:
    database: { host, port, user, password, dbname, log_lever, max_idle, max_open, max_lifetime, slow_threshold_ms }
    redis: { addr, password, db, pool_size, min_idle, conn_timeout }
    kafka: { brokers: ["127.0.0.1:9092"], topic: "seckill-order" }
  jwt: { secret: "change-me", expried: 3600, refresh_expried: 86400 }
  risk:
    enable: false
    login_rate: { rate: 5, burst: 10 }
    seckill_rate: { rate: 50, burst: 80 }
    pay_rate: { rate: 10, burst: 20 }
    product_rate: { rate: 1000, burst: 1000 }
    hotspot_burst: 100
  log: { level: "info", path: "log/api.log", max_age: 7, max_backups: 3, max_size: 100 }
  ```
- 启动：`go run ./cmd/api`（自动迁移模型），`go run ./cmd/worker`（消费 Kafka）；依赖 MySQL/Redis/Kafka，配置需与本地一致。
- 前端：在 `frontend/` 运行 `npm install`，`npm run dev` / `npm run build` / `npm run preview`；API 基址默认 `http://localhost:8000/api/v1`，可用 `VITE_API_BASE_URL` 覆盖。

## 后端开发规范
- 分层：handler -> service -> repository，仓储/服务提供 `WithContext`；统一响应 `{code,msg,data}`，错误码定义在 `internal/pkg/e`（包含 401/429/7xx 风控码）。
- 中间件：`middlerware` 内含 JWT、slog、Lua 令牌桶与黑/灰名单；`risk.enable` 开启后对登录/秒杀/支付/热点参数限流；静态上传通过 `/uploads` 暴露，CORS 已放通本地 5173。
- 业务要点：秒杀依赖 Redis Lua + Kafka，发送失败会回滚库存；订单/支付幂等在 `service/order.go`，支付回调校验状态；Snowflake 依赖 `server.machineid`，未配置会导致启动失败。
- 代码风格：`gofmt`/`goimports`；错误处理用 `fmt.Errorf(...%w...)` + `errors.Is`，不要新引入 `errors.Wrap`；避免跨层直接访问底层资源；日志用 `internal/pkg/logger`（slog+lumberjack）。

## 前端开发规范
- 入口与布局：`src/main.ts` 注册 Pinia/Router/Lenis；页面布局在 `layout/MainLayout.vue`。
- UI/样式：Tailwind 自定义 obsidian/magma 主题（`tailwind.config.js`、`assets/css/index.css`），基础组件在 `components/ui`（CVA 驱动），动效组件在 `components/motion`；优先复用现有组件/样式，不要重复引入 UI 库。
- 数据与状态：`lib/api.ts` 统一请求（校验 `code!=200` 抛错、自动 refresh token、本地键 `access_token`/`refresh_token`，含 `uploadImage`/`resolveAssetUrl`），组件或 store 通过该封装调用；已存在 store：`stores/userStore`、`stores/productStore`，其他页面可按需新增但保持相同模式。
- 路由：`router/index.ts` 管理，`meta.requiresAuth` 需依赖用户态；视图按功能存放在 `views/`（Auth/Home/Product/Orders/User）。
- 设计基调：暗夜黑金、玻璃态+流光动效（参考 `Home`/`Product` 页面），保持一致风格。

## 测试与验证
- 当前无自动化测试，改动核心逻辑（库存扣减、订单、限流、支付回调等）请补 Go 表驱动用例并确保 `go test ./...` 通过。
- 前端如新增复杂交互可引入/补充 Vitest + vue-test-utils（目前未配置），至少保证 `npm run build`/`vue-tsc -b` 通过。
- 压测需参考 `docs/perf.md` 与 `perf/` 脚本，必要时关闭/调低限流再跑 k6。

## 提交与安全
- Commit 信息用简洁中文动词短语，一次聚焦单主题；涉及接口/交互变更同步 `docs/backend-api.md` 或在 PR 说明，对照 `docs/frontend-plan.md`。
- 不要提交真实密钥/连接串；`config.yml`、`.env*`、`uploads/` 已忽略，上传文件仅用于本地调试。
