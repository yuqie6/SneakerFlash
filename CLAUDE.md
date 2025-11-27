# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SneakerFlash 是一个球鞋秒杀电商系统，采用前后端分离架构，专为高并发秒杀场景设计。

**技术栈：**
- 后端：Go 1.25 + Gin + GORM + Redis + Kafka
- 前端：Vue 3 + TypeScript + Vite + Tailwind CSS + Pinia
- 数据库：MySQL 8.4
- 消息队列：Kafka (KRaft 模式)
- 监控：Prometheus + Grafana

## Common Commands

### 后端
```bash
# 启动 HTTP API 服务器 (端口 8000)
go run ./cmd/api

# 启动 Kafka 消费 Worker
go run ./cmd/worker

# 运行测试
go test ./...
```

### 前端
```bash
cd frontend
npm install
npm run dev          # 开发服务器 (端口 5173)
npm run build        # 生产构建
npm run lint         # ESLint 检查
vue-tsc -b           # TypeScript 类型检查
```

### 基础设施
```bash
# 启动本地依赖 (MySQL/Redis/Kafka/Prometheus/Grafana)
docker-compose up -d
```

### 压测
```bash
k6 run -e BASE_URL=http://localhost:8000/api/v1 \
       -e RATE=200 -e DURATION=30s \
       -e USER_COUNT=1000 \
       perf/k6-seckill.js
```

## Architecture

### 后端分层架构

```
cmd/
├── api/main.go           # HTTP 服务入口
└── worker/main.go        # Kafka Consumer 入口

internal/
├── handler/              # HTTP 处理层 - 参数绑定、响应格式化
├── service/              # 业务逻辑层 - 核心业务、Redis/Kafka 操作
├── repository/           # 数据访问层 - GORM 数据库操作
├── model/                # 数据模型定义
├── middleware/           # JWT 验证、限流、日志、指标
├── config/               # Viper 配置管理
├── server/http.go        # Gin 路由注册
├── infra/                # Redis/Kafka 客户端初始化
└── pkg/                  # 工具包 (错误码、日志、JWT、Snowflake ID)
```

### 前端结构

```
frontend/src/
├── views/                # 页面组件
├── components/ui/        # Shadcn Vue 基础组件
├── composables/          # 可复用逻辑 (useSeckill, useCountDown)
├── stores/               # Pinia 状态管理
├── lib/api.ts            # Axios 实例 + Token 刷新拦截器
└── router/               # Vue Router 路由配置
```

### 秒杀核心流程

1. `POST /seckill` 请求到达
2. Redis Lua 脚本原子执行：检查库存 → 扣减库存 → 记录用户
3. 成功后发送消息到 Kafka，预写 pending 状态到 Redis
4. Worker 消费消息，落库订单和支付单
5. 前端轮询 `/orders/poll/{order_num}` 获取支付信息

### 关键设计模式

- **WithContext 模式**：Service/Repository 通过 `WithContext(ctx)` 传递 request_id
- **Lua 脚本原子性**：库存扣减和用户标记在 Redis 中原子执行
- **令牌桶限流**：基于 Redis Lua 脚本实现，支持接口级和参数级限流
- **幂等性**：order_num 唯一约束，支付回调幂等处理

## Configuration

配置文件：`./config.yml` 或环境变量 `SNEAKERFLASH_CONFIG` 指定路径

关键配置项：
- `server.machineid`：Snowflake ID 机器号（必需）
- `risk.enable`：风控开关
- `risk.*_rate`：各接口限流参数

## API Endpoints

基础路径：`http://localhost:8000/api/v1`

主要端点：
- 认证：`POST /register`, `POST /login`, `POST /refresh`
- 商品：`GET /products`, `GET /product/{id}`, `POST /products`
- 秒杀：`POST /seckill` (需认证 + 限流)
- 订单：`GET /orders`, `GET /orders/poll/{order_num}`
- 支付：`POST /payment/callback`

响应格式：`{ code, msg, data }`

## Error Codes

| 码值 | 含义 |
|------|------|
| 200 | 成功 |
| 401 | 未认证/Token 失效 |
| 10001 | 用户已存在 |
| 10002 | 用户不存在 |
| 20001 | 商品不存在 |
| 30001 | 秒杀售罄 |
| 30002 | 重复抢购 |
| 701 | 接口限流 |

## Frontend Design System

- 主题色：`obsidian-bg` (深黑), `magma` (橙金 #f97316)
- 动画：shimmer (流光), shake (震动), pulse-fast (脉搏)
- 玻璃态：`.glass` 类 (黑色半透明 + 毛玻璃)
- 组件库：Shadcn Vue

## Development Guidelines

- 后端错误处理使用 `fmt.Errorf("...: %w", err)` 包装
- 前端 API 调用统一使用 `lib/api.ts` 的 axios 实例
- 修改核心逻辑（库存、订单、限流）需补充 Go 表驱动测试
- 前端提交前需通过 `npm run build` 和 `vue-tsc -b` 验证
