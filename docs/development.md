# 开发与本地运行

## 适用对象
- 初次接手项目的开发者
- 需要恢复本地环境的维护者
- 调整后端链路、前端交互或文档体系的贡献者

## 环境要求
| 依赖 | 建议版本 | 用途 |
| --- | --- | --- |
| Go | `>= 1.25.4` | 后端 API / Worker |
| Node.js | `>= 18` | 前端开发与构建 |
| pnpm | `>= 10` | 前端依赖管理 |
| Docker | 最新稳定版 | 拉起 MySQL / Redis / Kafka / 监控 |
| golangci-lint | `v1.64+` | Go 代码检查 |

## 本地启动顺序
### 1. 准备配置
- 默认配置文件：仓库根 `config.yml`
- 或环境变量：`SNEAKERFLASH_CONFIG=/path/to/config.yml`
- 关键字段：
  - `server.port`
  - `server.machineid`
  - `data.database.*`
  - `data.redis.*`
  - `data.kafka.*`
  - `risk.enable`

详见 `configuration.md`

### 2. 启动依赖
```bash
docker compose up -d
```

### 3. 启动后端
```bash
go run ./cmd/api
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
make build-api
make build-worker
```

### Frontend
```bash
make lint-frontend
make frontend-build
```

### Swagger
```bash
swag init -g ./cmd/api/main.go -o ./docs
```

## 推荐开发流程
1. 阅读 `architecture.md` 和相关模块代码
2. 修改代码
3. 运行与改动范围最接近的验证命令
4. 执行 `golangci-lint run ./...` 与前端构建
5. 更新相关文档
6. 在 `governance.md` 记录重要文档变更

## 模块协作约束
### 后端
- 保持 `handler -> service -> repository` 分层边界
- 错误处理使用 `fmt.Errorf(... %w ...)`
- 避免跨层直接访问底层资源

### 前端
- API 统一走 `frontend/src/lib/api.ts`
- 共享状态优先放在 Pinia store
- 保持 Editorial 视觉风格一致：浅底纸张层次、硬边细边框、克制动效

## 代码检查
### Go
- 配置文件：`.golangci.yml`
- 默认启用：
  - `errcheck`
  - `gofmt`
  - `goimports`
  - `govet`
  - `ineffassign`
  - `misspell`
  - `staticcheck`
  - `unused`

### Frontend
- `frontend/package.json` 中提供 `pnpm lint`
- 构建即包含 `vue-tsc -b`
- 测试命令：`pnpm test:unit`、`pnpm test:e2e`

### 测试命令
- 后端单元测试：`make test` 或 `make test-unit`
- 后端集成测试：`make test-integration`
- 前端单元测试：`make test-frontend`
- 前端端到端测试：`make test-e2e`
- 全量验证：`make test-all`

## 当前测试现状
- 仓库当前未提交核心自动化测试
- 对秒杀、订单、支付、限流等核心逻辑进行改动时，建议优先补表驱动测试
- 压测与端到端演练目前主要依赖 `perf/` 脚本
- 详细测试设计、目录约定与分阶段落地计划见 `testing.md`

## 开发中最常见的问题
- Snowflake 初始化失败：缺少 `server.machineid`
- Docker 拉镜像失败：Docker daemon 仍绑定旧代理
- 秒杀永远 pending：Worker 未启动、Kafka 不通、Outbox 补偿未生效

更详细的排查手册见 `troubleshooting.md`
