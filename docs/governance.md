# 文档治理记录

## 目标
- 防止知识只存在于聊天记录、口头同步或个人脑内
- 保证 README、专题文档、Swagger 与代码行为尽量一致
- 为多人协作提供稳定的文档更新机制

## 文档分层
| 层级 | 文件 | 职责 |
| --- | --- | --- |
| L0 | `README.md` | 项目总览、快速启动、全局导航 |
| L1 | `docs/README.md` | 文档中心入口 |
| L2 | 专题文档 | 架构、开发、配置、运维、排障 |
| L3 | 契约文档 | `backend-api.md`、`swagger.*` |

## 更新规则
- 接口变化：更新 `docs/backend-api.md` 与 Swagger
- 配置变化：更新 `docs/configuration.md`
- 启动方式变化：更新 `README.md`、`docs/development.md`
- 运维方式变化：更新 `docs/operations.md`、`docs/troubleshooting.md`
- 重大结构变化：更新 `docs/architecture.md`

## 评审清单
- 文档是否有唯一入口
- 文件路径与命令是否可直接执行
- 是否与当前代码行为一致
- 是否覆盖了失败场景与排障动作
- 是否明确了责任边界

## 变更记录
| 日期 | 负责人 | 内容摘要 | 状态 |
| --- | --- | --- | --- |
| 2026-03-17 | Codex | 建立 README、文档中心、架构/开发/配置/运维/排障体系，并加入治理约束 | 完成 |
| 2026-03-17 | Codex | 将 Docker 旧代理问题沉淀为标准排障步骤 | 完成 |
| 2026-03-18 | Codex | 拆分开发/单机生产 Compose 与应用配置样例，并补充 Kafka 单机基线说明 | 完成 |
| 2026-03-18 | Codex | 统一 dev/prod/local/example 命名规则，收敛 Makefile 启动入口与实际读取流程 | 完成 |
| 2026-03-18 | Codex | 删除旧兼容配置入口并新增 make help，统一单人开发工作流 | 完成 |

## 待补事项
- 对齐 `docs/backend-api.md` 与 VIP / 优惠券实际实现
- 增加多节点部署文档与高可用 Kafka 基线
- 增加监控面板样例与告警阈值
- 增加核心链路测试策略说明
