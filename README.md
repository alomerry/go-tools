# go-tools

A comprehensive Go utility library providing various components, modules, and utilities for common development tasks.

[![go report](https://goreportcard.com/badge/github.com/alomerry/go-tools)](https://goreportcard.com/report/github.com/alomerry/go-tools)

## Requirements

- Go 1.25.0 and above

## Installation

```bash
go get github.com/alomerry/go-tools
```

## Overview

This library is organized into three main categories:

- **Components**: Reusable client wrappers for various services
- **Modules**: Standalone tools and utilities
- **Utils**: General-purpose utility functions

## Components

### Configuration Management

- **Apollo** (`components/apollo`): Apollo configuration center client with change listeners

### Databases

- **MySQL** (`components/mysql`): MySQL database client
- **MongoDB** (`components/mongo`): MongoDB client wrapper
- **Redis** (`components/redis`): Redis client with key generation utilities

### Message Queue & Streaming

- **Kafka** (`components/kafka`): Kafka client for topic management, message production, and metadata operations

### Object Storage (OSS)

- **OSS** (`components/oss`): Unified OSS client supporting multiple providers:
  - Qiniu Kodo
  - MinIO
  - AWS S3
  - Cloudflare R2

### Time Series Database

- **TSDB** (`components/tsdb`): Time series database client, currently supports InfluxDB

### Search & Analytics

- **Elasticsearch** (`components/es`): Elasticsearch typed client wrapper

### Infrastructure

- **Kubernetes** (`components/k8s`): Kubernetes client for managing deployments, pods, services, and resources
- **gRPC** (`components/grpc`): gRPC utilities including custom header matchers

### Monitoring & Logging

- **Monitor** (`components/monitor`): System monitoring with CPU, memory, disk, and network statistics
- **Log** (`components/log`): Logrus formatter and logging utilities

## Modules

### DNS Tools

- **DNS** (`modules/dns`): DNS management tools supporting:
  - Alibaba Cloud DNS (AliDNS)
  - Cloudflare DNS

### File Management

- **Pusher** (`modules/pusher`): File upload tool for OSS with support for:
  - File existence checking
  - Automatic upload on file changes
  - Cloudflare R2 support

### Excel Processing

- **SGS** (`modules/sgs`): Excel processing tools for delay analysis and reporting

## Utils

### Data Structures

- **Algorithm** (`utils/algorithm`):
  - Binary Search Tree (BST)
  - Queue
  - Set (generic implementation)

### Database Utilities

- **DB** (`utils/db`): Database backup tools
  - MySQL dump functionality
  - MongoDB ObjectID utilities

### File Operations

- **Files** (`utils/files`): File manipulation utilities
- **Tar** (`utils/tar`): TAR archive operations
- **Zip** (`utils/zip`): ZIP archive operations

### Data Processing

- **JSON** (`utils/json`): JSON processing utilities
- **Array** (`utils/array`): Array manipulation functions
- **Maps** (`utils/maps`): Concurrent map implementations
- **String** (`utils/string`): String utility functions
- **Random** (`utils/random`): Random string generation

### Network & Web

- **Net** (`utils/net`): Network utilities
- **UA** (`utils/ua`): User-Agent parsing utilities

### Security & Authentication

- **JWT** (`utils/jwt`): JWT token generation and validation

### Time & Context

- **Time** (`utils/time`): Time utility functions
- **Context** (`utils/context.go`): Context utilities

### Other Utilities

- **Base** (`utils/base`): Base utility functions
- **Vars** (`utils/vars`): Variable utilities
- **Func** (`utils/func.go`): Function utilities
- **Struct** (`utils/struct.go`): Struct reflection utilities

## Static

The `static` directory contains configuration constants, environment variable helpers, and error definitions:

- **Cons** (`static/cons`): Application constants
- **Env** (`static/env`): Environment variable helpers
- **Errors** (`static/errors`): Error definitions

## Usage Examples

### OSS Client

```go
import "github.com/alomerry/go-tools/components/oss"
import "github.com/alomerry/go-tools/components/oss/meta"

cfg := meta.Config{
    Type:   meta.ClientTypeR2,
    Bucket: "my-bucket",
    // ... other config
}

client, err := oss.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}
```

### Kafka Client

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

### System Monitor

```go
import "github.com/alomerry/go-tools/components/monitor"

monitor := monitor.NewSystemMonitor(
    monitor.WithContext(ctx),
    monitor.WithInterval(30*time.Second),
    monitor.WithCallback(func(stats *monitor.SystemStats) error {
        log.Printf("CPU: %.2f%%, Memory: %.2f%%", 
            stats.CPUUsage, stats.MemoryUsage)
        return nil
    }),
)

monitor.Watch()
```

### Database Dump

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

### Set Operations

```go
import "github.com/alomerry/go-tools/utils/algorithm"

set := algorithm.Instance[string]()
set.Insert("a").Insert("b").Insert("c")

if set.Has("a") {
    fmt.Println("Set contains 'a'")
}

items := set.ToArray()
```

## License

See [LICENSE](LICENSE) file for details.

## Thanks

Thanks for free JetBrains Open Source license

<a href="https://www.jetbrains.com/?from=alomerry/go-tools" target="_blank">
<img src="https://user-images.githubusercontent.com/1787798/69898077-4f4e3d00-138f-11ea-81f9-96fb7c49da89.png" height="100"/></a>
