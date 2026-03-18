# SneakerFlash 文档中心

> 这是仓库内的一级文档入口。建议阅读顺序：`README.md` -> `docs/README.md` -> 对应专题文档。

## 文档地图
### 总览
- `index.md`：文档目录索引
- `plan.md`：阶段路线图与当前完成度审计
- `governance.md`：文档治理规则、更新记录与维护责任

### 架构与设计
- `architecture.md`：系统分层、秒杀时序、最终一致性与风控设计
- `frontend-plan.md`：前端当前实现方案、路由与视觉约束

### 开发与配置
- `development.md`：本地开发、运行、lint、构建与协作流程
- `testing.md`：当前测试覆盖、运行方式与补强建议
- `configuration.md`：配置项说明、环境变量与常见取值建议

### 运维与排障
- `operations.md`：Docker Compose、监控、压测、日常运维操作
- `troubleshooting.md`：常见故障与标准化排查手册
- `perf.md`：压测使用说明

### 接口与契约
- `backend-api.md`：当前后端接口摘要
- `swagger.yaml`：Swagger 规范
- `swagger.json`：Swagger JSON 产物
- `docs.go`：Swagger 生成代码

### 归档
- `archive/README.md`：归档文档索引

## 阅读建议
### 新人恢复上下文
1. `README.md`
2. `architecture.md`
3. `development.md`
4. `backend-api.md`
5. `plan.md`

### 开发前必读
1. `development.md`
2. `configuration.md`
3. `backend-api.md`
4. `frontend-plan.md`
5. `testing.md`

### 运维排障
1. `operations.md`
2. `troubleshooting.md`
3. `perf.md`

## 文档维护规则
- 接口变化：更新 `backend-api.md` 与 Swagger 产物
- 配置变化：更新 `configuration.md`
- 启动/运维变化：更新 `operations.md` 与 `troubleshooting.md`
- 前端结构/交互变化：更新 `frontend-plan.md`
- 大版本演进或阶段完成度变化：更新 `plan.md` 与 `governance.md`
- 失效但需要保留背景的文档：移动到 `docs/archive/`
