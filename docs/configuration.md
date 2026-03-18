# 配置说明

## 配置来源
- 开发模板：`config.dev.yml.example`
- 单机生产模板：`config.prod.yml.example`
- 开发实际文件：`config.dev.local.yml`
- 单机生产实际文件：`config.prod.local.yml`
- 覆盖方式：环境变量 `SNEAKERFLASH_*`
- 指定文件：`SNEAKERFLASH_CONFIG=/path/to/config.local.yml`

## 读取规则
1. 必须设置 `SNEAKERFLASH_CONFIG`
2. 程序读取该路径指向的 yml 文件
3. 读取完成后，再应用 `SNEAKERFLASH_` 前缀的环境变量覆盖同名配置

> 当前项目不再依赖默认 `config.yml`，统一通过 `make` 显式传入 `config.<env>.local.yml`。

例如：
- `SNEAKERFLASH_CONFIG=./config.dev.local.yml`
- `SNEAKERFLASH_SERVER_PORT=0.0.0.0:8001`
- `SNEAKERFLASH_DATA_KAFKA_TOPIC=seckill_orders`
- `SNEAKERFLASH_DATA_KAFKA_CONSUMER_GROUP=sneaker-group-dev`

## 推荐结构
### 开发环境
```yaml
server:
  port: "0.0.0.0:8000"
  machineid: 1
  upload_dir: "uploads"

data:
  database:
    host: "127.0.0.1"
    port: 13306
    user: "sneaker"
    password: "sneaker_dev"
    dbname: "sneaker_flash"
    log_lever: 2
    max_idle: 10
    max_open: 100
    max_lifetime: 3600
    max_idle_time: 300
    slow_threshold_ms: 500
  redis:
    addr: "127.0.0.1:16379"
    password: "123456"
    db: 0
    pool_size: 100
    min_idle: 20
    conn_timeout: 5
  kafka:
    brokers: ["127.0.0.1:19092"]
    topic: "seckill_orders"
    consumer_group: "sneaker-group-dev"
    initial_offset: "oldest"
    batch_size: 100
    flush_interval: 200
    max_retries: 3
    dlq_topic: "seckill-order-dlq"
    outbox_scan_interval: 30
    outbox_timeout: 60

jwt:
  secret: "change-me-dev"
  expried: 86400
  refresh_expried: 604800

risk:
  enable: false
  login_rate: { rate: 5, burst: 10 }
  seckill_rate: { rate: 50, burst: 80 }
  pay_rate: { rate: 10, burst: 20 }
  product_rate: { rate: 1000, burst: 1000 }
  hotspot_burst: 100

log:
  level: "debug"
  path: "./log/app"
  max_age: 7
  max_backups: 30
  max_size: 100
```

### 单机生产基线
```yaml
server:
  port: "127.0.0.1:8000"
  machineid: 1
  upload_dir: "/var/lib/sneakerflash/uploads"

data:
  database:
    host: "127.0.0.1"
    port: 3306
    user: "sneaker"
    password: "<replace-with-secret>"
    dbname: "sneaker_flash"
    log_lever: 1
    max_idle: 20
    max_open: 200
    max_lifetime: 1800
    max_idle_time: 300
    slow_threshold_ms: 300
  redis:
    addr: "127.0.0.1:6379"
    password: "<replace-with-secret>"
    db: 0
    pool_size: 200
    min_idle: 50
    conn_timeout: 3
  kafka:
    brokers: ["127.0.0.1:9092"]
    topic: "seckill_orders"
    consumer_group: "sneaker-group-prod"
    initial_offset: "oldest"
    batch_size: 200
    flush_interval: 100
    max_retries: 5
    dlq_topic: "seckill-order-dlq"
    outbox_scan_interval: 10
    outbox_timeout: 30

jwt:
  secret: "<replace-with-secret>"
  expried: 3600
  refresh_expried: 86400

risk:
  enable: true
  login_rate: { rate: 50, burst: 100 }
  seckill_rate: { rate: 1500, burst: 3000 }
  pay_rate: { rate: 100, burst: 200 }
  product_rate: { rate: 1000, burst: 1000 }
  hotspot_burst: 100

log:
  level: "info"
  path: "/var/log/sneakerflash/app.log"
  max_age: 30
  max_backups: 10
  max_size: 100
```

## 关键字段说明
### `server`
| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `port` | 是 | API 监听地址 |
| `machineid` | 是 | Snowflake 机器号 |
| `upload_dir` | 否 | 上传文件目录 |

### `data.database`
| 字段 | 说明 |
| --- | --- |
| `host` / `port` | MySQL 地址 |
| `user` / `password` | 数据库账号 |
| `dbname` | 数据库名 |
| `log_lever` | GORM 日志级别 |
| `max_idle` / `max_open` | 连接池参数 |
| `max_lifetime` / `max_idle_time` | 连接生命周期 |
| `slow_threshold_ms` | 慢查询阈值 |

### `data.redis`
| 字段 | 说明 |
| --- | --- |
| `addr` | Redis 地址 |
| `password` | Redis 密码 |
| `db` | DB 编号 |
| `pool_size` | 连接池大小 |
| `min_idle` | 最小空闲连接数 |
| `conn_timeout` | 连接超时 |

### `data.kafka`
| 字段 | 说明 |
| --- | --- |
| `brokers` | Broker 列表 |
| `topic` | 秒杀主题 |
| `consumer_group` | Worker 消费组 ID |
| `initial_offset` | 消费组首次启动时的起始位点，支持 `oldest` / `newest` |
| `batch_size` | 批量消费数量 |
| `flush_interval` | 批量聚合等待时间 |
| `max_retries` | 最大重试次数 |
| `dlq_topic` | 死信主题 |
| `outbox_scan_interval` | Outbox 扫描周期 |
| `outbox_timeout` | Outbox 超时阈值 |

### `jwt`
- `expried`：access token 过期秒数
- `refresh_expried`：refresh token 过期秒数

### `risk`
- `enable`：总开关
- `*_rate.rate`：每秒令牌数
- `*_rate.burst`：桶容量
- `hotspot_burst`：热点参数突发量

## 环境变量映射
- 使用 `SNEAKERFLASH_` 前缀
- 点号转下划线
- 例如：
  - `SNEAKERFLASH_SERVER_PORT`
  - `SNEAKERFLASH_DATA_DATABASE_HOST`
  - `SNEAKERFLASH_RISK_ENABLE`
  - `SNEAKERFLASH_DATA_KAFKA_CONSUMER_GROUP`
  - `SNEAKERFLASH_DATA_KAFKA_INITIAL_OFFSET`

## 推荐配置建议
### 本地开发
- `risk.enable=false`
- Kafka broker 使用 `127.0.0.1:19092`
- 日志级别使用 `info` 或 `debug`
- 配套编排文件使用 `docker-compose.dev.yaml`
- 推荐通过 `make dev-init` 生成 `.env.dev.local` 和 `config.dev.local.yml`

### 压测环境
- 适当调高 Redis、DB 连接池
- 适当放宽限流参数
- 单独观察 Kafka lag 与 DB 慢查询

### 生产环境
- 禁止把真实密钥提交到仓库
- Secret 放环境变量或密钥系统
- 单机基线默认使用 `127.0.0.1:9092`
- 明确区分 dev / stage / prod 配置
- `docker-compose.prod.yaml` 只解决单机部署规范，不解决 Kafka 高可用
- 推荐通过 `make prod-init` 生成 `.env.prod.local` 和 `config.prod.local.yml`
