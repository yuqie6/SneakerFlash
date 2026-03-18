# SneakerFlash

> 面向高并发球鞋秒杀场景的全栈项目。后端使用 `Go + Gin + GORM + Redis + Kafka` 构建秒杀、异步建单、支付回调、VIP/优惠券闭环；前端使用 `Vue 3 + TypeScript + Vite + Tailwind` 提供 Editorial 风格的交易体验。

## 项目定位
- **核心场景**：限量商品秒杀、异步建单、订单支付、VIP 成长体系、优惠券发放与核销。
- **系统目标**：高并发入口抗压、库存一致性、消息最终一致性、可观测与可运维。
- **当前形态**：单仓库前后端项目，适合本地开发、压测演练与架构演进。

## 快速导航
- **文档中心**：`docs/README.md`
- **架构总览**：`docs/architecture.md`
- **开发指南**：`docs/development.md`
- **测试方案**：`docs/testing.md`
- **运维与排障**：`docs/operations.md`
- **配置说明**：`docs/configuration.md`
- **故障排查**：`docs/troubleshooting.md`
- **接口契约**：`docs/backend-api.md`、`docs/swagger.yaml`
- **前端方案**：`docs/frontend-plan.md`
- **阶段路线图**：`docs/plan.md`

## 技术栈
- **后端**：`Go 1.25`、`Gin`、`GORM`、`Redis`、`Kafka`、`Viper`
- **前端**：`Vue 3`、`TypeScript`、`Vite`、`Pinia`、`Vue Router`、`Tailwind CSS`
- **基础设施**：`MySQL 8.4`、`Redis 8`、`Kafka 4.x (KRaft)`、`Prometheus`、`Grafana`
- **工程工具**：`golangci-lint`、`ESLint`、`Prettier`、`k6`

## 仓库结构
```text
.
├── cmd/                     # API / Worker 入口
├── internal/                # 后端业务、仓储、基础设施、中间件
├── frontend/                # Vue3 前端
├── docs/                    # 项目文档中心
├── perf/                    # 压测脚本与辅助工具
├── docker/                  # Prometheus / Grafana 配置
├── docker-compose.yaml      # 本地依赖编排
├── Makefile                 # 常用开发命令
└── .golangci.yml            # Go 代码检查配置
```

## 环境准备
### 必需依赖
- `Go >= 1.25.4`
- `Node.js >= 18`
- `Docker + Docker Compose`
- `MySQL / Redis / Kafka`（可由 `docker compose` 拉起）

### 本地配置
- 默认读取仓库根 `config.yml`
- 或设置环境变量 `SNEAKERFLASH_CONFIG=/path/to/config.yml`
- `server.machineid` 必填，否则 Snowflake 初始化失败

配置详情见 `docs/configuration.md`

## 本地启动
### 1. 启动依赖
```bash
docker compose up -d
```

### 2. 启动后端 API
```bash
go run ./cmd/api
```

### 3. 启动 Worker
```bash
go run ./cmd/worker
```

### 4. 启动前端
```bash
cd frontend
pnpm install
pnpm dev
```

## 常用命令
### Go
```bash
make lint-go
make test
make test-integration
make build-api
make build-worker
```

### Frontend
```bash
make lint-frontend
make test-frontend
make test-e2e
make frontend-build
```

### Full Check
```bash
make lint
make test-all
```

## 核心业务链路
1. 用户请求 `POST /api/v1/seckill`
2. Redis Lua 原子完成防重复购买与库存预扣
3. 服务写入 Outbox，并异步投递 Kafka
4. Worker 批量消费消息，落库订单与支付单
5. 前端通过 `GET /api/v1/orders/poll/:order_num` 轮询异步结果
6. 支付回调推动订单状态、成长等级与优惠券状态变更

架构细节见 `docs/architecture.md`

## 工程约束
- Go 代码检查使用 `.golangci.yml`
- 文档入口为 `README.md` 与 `docs/README.md`
- 接口、配置、运维方式变更必须同步更新文档
- 当前未提交自动化测试，核心逻辑变更建议优先补表驱动测试

## 当前已知事项
- Docker 若仍绑定旧代理，拉镜像会报 `proxyconnect tcp: dial tcp 127.0.0.1:10808: connect: connection refused`
- 修复步骤见 `docs/troubleshooting.md`
- `docs/backend-api.md` 与 Swagger 产物仍需持续对齐 VIP / 优惠券最新能力

## 后续建议
- 补齐 `internal/service` 层核心测试
- 补齐接口文档与 Swagger 的一致性检查
- 为运维文档增加监控面板截图与告警阈值基线
