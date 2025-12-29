# go-tools

一个全面的 Go 工具库，提供各种组件、模块和实用工具，用于常见的开发任务。

[![go report](https://goreportcard.com/badge/github.com/alomerry/go-tools)](https://goreportcard.com/report/github.com/alomerry/go-tools)

## 要求

- Go 1.25.0 及以上版本

## 安装

```bash
go get github.com/alomerry/go-tools
```

## 概述

该库主要分为三个类别：

- **Components（组件）**：各种服务的可复用客户端封装
- **Modules（模块）**：独立的工具和实用程序
- **Utils（工具）**：通用工具函数

## Components（组件）

### 配置管理

- **Apollo** (`components/apollo`): Apollo 配置中心客户端，支持配置变更监听

### 数据库

- **MySQL** (`components/mysql`): MySQL 数据库客户端
- **MongoDB** (`components/mongo`): MongoDB 客户端封装
- **Redis** (`components/redis`): Redis 客户端，包含键生成工具

### 消息队列与流处理

- **Kafka** (`components/kafka`): Kafka 客户端，支持主题管理、消息生产和元数据操作

### 对象存储 (OSS)

- **OSS** (`components/oss`): 统一的对象存储客户端，支持多种提供商：
  - 七牛云 Kodo
  - MinIO
  - AWS S3
  - Cloudflare R2

### 时序数据库

- **TSDB** (`components/tsdb`): 时序数据库客户端，目前支持 InfluxDB

### 搜索与分析

- **Elasticsearch** (`components/es`): Elasticsearch 类型化客户端封装

### 基础设施

- **Kubernetes** (`components/k8s`): Kubernetes 客户端，用于管理部署、Pod、服务和资源
- **gRPC** (`components/grpc`): gRPC 工具，包括自定义头部匹配器

### 监控与日志

- **Monitor** (`components/monitor`): 系统监控，包含 CPU、内存、磁盘和网络统计信息
- **Log** (`components/log`): Logrus 格式化器和日志工具

## Modules（模块）

### DNS 工具

- **DNS** (`modules/dns`): DNS 管理工具，支持：
  - 阿里云 DNS (AliDNS)
  - Cloudflare DNS

### 文件管理

- **Pusher** (`modules/pusher`): 文件上传工具，支持 OSS，功能包括：
  - 文件存在性检查
  - 文件变更时自动上传
  - Cloudflare R2 支持

### Excel 处理

- **SGS** (`modules/sgs`): Excel 处理工具，用于延迟分析和报告生成

## Utils（工具）

### 数据结构

- **Algorithm** (`utils/algorithm`):
  - 二叉搜索树 (BST)
  - 队列
  - 集合（泛型实现）

### 数据库工具

- **DB** (`utils/db`): 数据库备份工具
  - MySQL 转储功能
  - MongoDB ObjectID 工具

### 文件操作

- **Files** (`utils/files`): 文件操作工具
- **Tar** (`utils/tar`): TAR 归档操作
- **Zip** (`utils/zip`): ZIP 归档操作

### 数据处理

- **JSON** (`utils/json`): JSON 处理工具
- **Array** (`utils/array`): 数组操作函数
- **Maps** (`utils/maps`): 并发 Map 实现
- **String** (`utils/string`): 字符串工具函数
- **Random** (`utils/random`): 随机字符串生成

### 网络与 Web

- **Net** (`utils/net`): 网络工具
- **UA** (`utils/ua`): User-Agent 解析工具

### 安全与认证

- **JWT** (`utils/jwt`): JWT 令牌生成和验证

### 时间与上下文

- **Time** (`utils/time`): 时间工具函数
- **Context** (`utils/context.go`): 上下文工具

### 其他工具

- **Base** (`utils/base`): 基础工具函数
- **Vars** (`utils/vars`): 变量工具
- **Func** (`utils/func.go`): 函数工具
- **Struct** (`utils/struct.go`): 结构体反射工具

## Static（静态配置）

`static` 目录包含配置常量、环境变量辅助函数和错误定义：

- **Cons** (`static/cons`): 应用常量
- **Env** (`static/env`): 环境变量辅助函数
- **Errors** (`static/errors`): 错误定义

## 使用示例

### OSS 客户端

```go
import "github.com/alomerry/go-tools/components/oss"
import "github.com/alomerry/go-tools/components/oss/meta"

cfg := meta.Config{
    Type:   meta.ClientTypeR2,
    Bucket: "my-bucket",
    // ... 其他配置
}

client, err := oss.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}
```

### Kafka 客户端

```go
import "github.com/alomerry/go-tools/components/kafka"

client := kafka.NewKafkaClient(ctx,
    kafka.WithAddresses("localhost:9092"),
)

topics, err := client.ListTopics()
if err != nil {
    log.Fatal(err)
}
```

### 系统监控

```go
import "github.com/alomerry/go-tools/components/monitor"

monitor := monitor.NewSystemMonitor(
    monitor.WithContext(ctx),
    monitor.WithInterval(30*time.Second),
    monitor.WithCallback(func(stats *monitor.SystemStats) error {
        log.Printf("CPU: %.2f%%, 内存: %.2f%%", 
            stats.CPUUsage, stats.MemoryUsage)
        return nil
    }),
)

monitor.Watch()
```

### 数据库备份

```go
import "github.com/alomerry/go-tools/utils/db"
import "github.com/alomerry/go-tools/static/cons"

tool := db.NewDumpTool(
    db.MySQLDumpCmdParam("user:pass@tcp(localhost:3306)/dbname"),
    db.SetDumpPath("/tmp/backups"),
)

files, err := tool.DumpDbs(cons.Database{
    Type: cons.MySQL,
    Name: "mydb",
})
```

### 集合操作

```go
import "github.com/alomerry/go-tools/utils/algorithm"

set := algorithm.Instance[string]()
set.Insert("a").Insert("b").Insert("c")

if set.Has("a") {
    fmt.Println("集合包含 'a'")
}

items := set.ToArray()
```

## 许可证

详情请参阅 [LICENSE](LICENSE) 文件。

## 致谢

感谢 JetBrains 提供的免费开源许可证

<a href="https://www.jetbrains.com/?from=alomerry/go-tools" target="_blank">
<img src="https://user-images.githubusercontent.com/1787798/69898077-4f4e3d00-138f-11ea-81f9-96fb7c49da89.png" height="100"/></a>

