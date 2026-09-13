# go-tools

一个全面的 Go 工具库，提供各种组件与实用工具，用于常见的后端开发任务。

[![go report](https://goreportcard.com/badge/github.com/alomerry/go-tools)](https://goreportcard.com/report/github.com/alomerry/go-tools)
[![CI](https://github.com/alomerry/go-tools/actions/workflows/ci.yml/badge.svg)](https://github.com/alomerry/go-tools/actions/workflows/ci.yml)

## 要求

- Go 1.25.0 及以上版本

## 安装

```bash
go get github.com/alomerry/go-tools
```

## 概述

该库主要分为四类：

- **Components（组件）**：各类中间件与外部服务的可复用客户端封装
- **Utils（工具）**：通用工具函数与数据结构
- **Static（静态）**：配置常量、环境变量辅助与错误定义
- **其他**：analysis 自定义 vet 工具、model 公共模型、test 测试辅助

## Components（组件）

### 配置管理

- **Apollo** (`components/apollo`): Apollo 配置中心客户端，支持配置变更监听与 `Dynamic[T]` 泛型热更新快照；`mananger` 子包、`sdk` OpenAPI 子包

### 数据库

- **MySQL** (`components/mysql`): MySQL 客户端（bun ORM 封装）
- **MongoDB** (`components/mongo`): MongoDB 客户端封装
- **Redis** (`components/redis`): Redis 客户端，包含键生成工具

### 消息队列

- **Kafka** (`components/kafka`): Kafka 客户端，支持主题管理、消息生产（Producer/AdminClient）

### 对象存储 (OSS)

- **OSS** (`components/oss`): 统一的对象存储客户端，支持多种提供商：
  - 七牛云 Kodo
  - MinIO
  - AWS S3（经 S3 兼容协议）
  - Cloudflare R2
  - RustFS

### 时序数据库

- **TSDB** (`components/tsdb`): 时序数据库客户端，目前支持 InfluxDB（经 Kafka 异步写入链路），带查询构造与并发安全封装

### 搜索

- **Elasticsearch** (`components/es`): Elasticsearch 客户端封装，`sdk` 子包

### 服务打点与可观测

- **CAT** (`components/cat`): CAT 风格服务端打点（Transaction/Event/Problem），点位经 Kafka 异步落 InfluxDB
- **Monitor** (`components/monitor`): 系统监控（host/docker 类别），CPU、内存、磁盘、网络与负载统计
- **Log** (`components/log`): Logrus 格式化器和日志工具（trace id 注入、单行化）

### 通知

- **Notify** (`components/notify`): 统一通知管理器，驱动化设计：
  - Bark (`drivers/bark`)
  - Console (`drivers/console`)
  - Kook (`drivers/kook`，卡片消息)
  - Kook Webhook 加密回调 (`components/kook/webhook`)

### 基础设施

- **Kubernetes** (`components/k8s`): Kubernetes 客户端，管理 Deployment、Pod、Service 与任意 YAML 资源；含 cAdvisor Pod 用量采集（apiserver proxy）
- **Tekton** (`components/tekton`): Tekton 客户端（PipelineRun/TaskRun/日志）
- **SSH** (`components/ssh`): SSH 客户端，支持私钥/密码认证与会话执行，可选主机密钥校验
- **gRPC** (`components/grpc`): gRPC 工具（自定义头部匹配器）

### HTTP

- **HTTP** (`components/http`): resty 封装的 HTTP 客户端（重试、CAT 事务集成）
- **Ext** (`components/ext`): 扩展加载框架（`LoadExt`），统一初始化 Apollo/MySQL/Mongo/Redis/Metric/Logger 等

### 其他

- **Cleaner** (`components/cleaner`): 目录清理工具
- **Collect** (`components/collect`): 监控采集 Agent（本机采集 + SSH 批量管理）

## Utils（工具）

### 数据结构

- **Algorithm** (`utils/algorithm`): 队列、集合（泛型实现）
- **Cache** (`utils/cache`): 泛型 LRU 缓存（并发安全、TTL 支持）
- **Maps** (`utils/maps`): 并发安全 Map

### 数据库工具

- **DB** (`utils/db`): 数据库备份工具（mysqldump 封装，密码经环境变量传递）

### 文件操作

- **Files** (`utils/files`): 文件操作工具
- **Tar** (`utils/tar`): TAR 归档解压（含 Zip Slip 防护）
- **Zip** (`utils/zip`): ZIP 归档操作

### 数据处理

- **JSON** (`utils/json`): JSON 处理工具
- **Array** (`utils/array`): 数组操作函数
- **String** (`utils/string`): 字符串工具函数
- **Random** (`utils/random`): 随机字符串生成（非密码学安全）

### 网络与 Web

- **Net** (`utils/net`): 网络工具
- **UA** (`utils/ua`): User-Agent 解析工具
- **Resty** (`utils/resty`): resty 请求中间件（CAT 事务、重试条件）

### 安全与认证

- **JWT** (`utils/jwt`): JWT 令牌生成与验证（限定 HS256）
- **Crypto** (`utils/crypto`): AES-256-CBC 加解密与 PKCS5 填充

### 时间与上下文

- **Time** (`utils/time`): 时间工具函数
- **Trace** (`utils/trace`): trace id 上下文工具
- **Proto** (`utils/proto`): protobuf 工具
- **Shell** (`utils/shell`): shell 命令执行

### 其他工具

- **Base** (`utils/base`): 基础工具函数
- **Vars** (`utils/vars`): 变量工具
- **Struct** (`utils/struct.go`): 结构体反射调用工具

## Static（静态配置）

`static` 目录包含配置常量、环境变量辅助函数和错误定义：

- **Cons** (`static/cons`): 应用常量
- **Env** (`static/env`): 环境变量读取辅助
- **Errors** (`static/errors`): 统一错误定义

## Model（公共模型）

`model` 目录包含跨包共享的数据模型（监控统计、OSS 配置等）。

## Analysis（自定义 vet）

`analysis` 包含基于 go/analysis 的自定义静态检查（如 `redis.Generator.GenKey` 参数必须为 cons 包常量），`analysis/cmd` 为独立构建入口。

## Test（测试辅助）

`test` 包含测试套件辅助（`BaseSuite`）与集成测试开关（`GO_TOOLS_TEST_INTEGRATION=1` 启用外部环境依赖测试）。

## 开发

```bash
# 构建
go build ./...

# 静态检查
go vet ./...

# 测试（-short 跳过 demo 长跑测试；外部环境依赖测试默认 skip）
go test -short ./...

# 本地 lint（如安装了 golangci-lint）
golangci-lint run
```

## License

[MIT](LICENSE)
