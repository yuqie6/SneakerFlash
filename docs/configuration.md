# 配置说明

## 配置来源
- 默认文件：`config.yml`
- 覆盖方式：环境变量 `SNEAKERFLASH_*`
- 指定文件：`SNEAKERFLASH_CONFIG=/path/to/config.yml`

## 推荐结构
```yaml
server:
  port: ":8000"
  machineid: 1
  upload_dir: "uploads"

data:
  database:
    host: "127.0.0.1"
    port: 3306
    user: "root"
    password: "root"
    dbname: "sneaker_flash"
    log_lever: 3
    max_idle: 10
    max_open: 100
    max_lifetime: 300
    max_idle_time: 300
    slow_threshold_ms: 200
  redis:
    addr: "127.0.0.1:6379"
    password: "123456"
    db: 0
    pool_size: 50
    min_idle: 10
    conn_timeout: 5
  kafka:
    brokers: ["127.0.0.1:9092"]
    topic: "seckill-order"
    batch_size: 100
    flush_interval: 200
    max_retries: 3
    dlq_topic: "seckill-order-dlq"
    outbox_scan_interval: 30
    outbox_timeout: 60

jwt:
  secret: "change-me"
  expried: 3600
  refresh_expried: 86400

risk:
  enable: false
  login_rate: { rate: 5, burst: 10 }
  seckill_rate: { rate: 50, burst: 80 }
  pay_rate: { rate: 10, burst: 20 }
  product_rate: { rate: 1000, burst: 1000 }
  hotspot_burst: 100

log:
  level: "info"
  path: "log/api.log"
  max_age: 7
  max_backups: 3
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

## 推荐配置建议
### 本地开发
- `risk.enable=false`
- Kafka broker 使用本地地址
- 日志级别使用 `info` 或 `debug`

### 压测环境
- 适当调高 Redis、DB 连接池
- 适当放宽限流参数
- 单独观察 Kafka lag 与 DB 慢查询

### 生产环境
- 禁止把真实密钥提交到仓库
- Secret 放环境变量或密钥系统
- 明确区分 dev / stage / prod 配置

