# go-tools

A comprehensive Go utility library providing various components, modules, and utilities for common development tasks.

[![go report](https://goreportcard.com/badge/github.com/alomerry/go-tools)](https://goreportcard.com/report/github.com/alomerry/go-tools)
[![CI](https://github.com/alomerry/go-tools/actions/workflows/ci.yml/badge.svg)](https://github.com/alomerry/go-tools/actions/workflows/ci.yml)

## Requirements

- Go 1.25.0 and above

## Installation

```bash
go get github.com/alomerry/go-tools
```

## Overview

This library is organized into four main categories:

- **Components**: reusable client wrappers for middleware and external services
- **Utils**: general-purpose utility functions and data structures
- **Static**: configuration constants, environment helpers, and error definitions
- **Misc**: custom go/analysis vet tool, shared models, and test helpers

> 中文文档见 [README_ZH.md](README_ZH.md)。

## Components

### Configuration Management

- **Apollo** (`components/apollo`): Apollo config center client with change listeners and generic `Dynamic[T]` hot-reload snapshots; includes `mananger` and `sdk` (OpenAPI) sub-packages

### Databases

- **MySQL** (`components/mysql`): MySQL database client (bun ORM wrapper)
- **MongoDB** (`components/mongo`): MongoDB client wrapper
- **Redis** (`components/redis`): Redis client with key generation utilities

### Message Queue

- **Kafka** (`components/kafka`): Kafka client for topic management, message production, and admin operations

### Object Storage (OSS)

- **OSS** (`components/oss`): unified OSS client supporting multiple providers:
  - Qiniu Kodo
  - MinIO
  - AWS S3 (S3-compatible protocol)
  - Cloudflare R2
  - RustFS

### Time Series Database

- **TSDB** (`components/tsdb`): time series database client, currently supports InfluxDB (async write pipeline via Kafka), with query builder and concurrency-safe wrappers

### Search

- **Elasticsearch** (`components/es`): Elasticsearch client wrapper with `sdk` sub-package

### Metrics & Observability

- **CAT** (`components/cat`): CAT-style server-side instrumentation (Transaction/Event/Problem), points land in InfluxDB via Kafka
- **Monitor** (`components/monitor`): system monitoring (host/docker categories) with CPU, memory, disk, network, and load statistics
- **Log** (`components/log`): logrus formatter and logging utilities (trace id injection, single-line output)

### Notifications

- **Notify** (`components/notify`): unified notification manager with a driver-based design:
  - Bark (`drivers/bark`)
  - Console (`drivers/console`)
  - Kook (`drivers/kook`, card messages)
  - Kook webhook encryption (`components/kook/webhook`)

### Infrastructure

- **Kubernetes** (`components/k8s`): Kubernetes client for managing deployments, pods, services, and arbitrary YAML resources; includes cAdvisor pod usage collection via apiserver proxy
- **Tekton** (`components/tekton`): Tekton client (PipelineRun/TaskRun/logs)
- **SSH** (`components/ssh`): SSH client with private key/password auth and optional host key verification
- **gRPC** (`components/grpc`): gRPC utilities including custom header matchers

### HTTP

- **HTTP** (`components/http`): resty-wrapped HTTP client (retry, CAT transaction integration)
- **Ext** (`components/ext`): extension loading framework (`LoadExt`) that bootstraps Apollo/MySQL/Mongo/Redis/Metric/Logger uniformly

### Misc

- **Cleaner** (`components/cleaner`): directory cleaning utility
- **Collect** (`components/collect`): monitoring collection agent (local collection + SSH fleet management)

## Utils

### Data Structures

- **Algorithm** (`utils/algorithm`): queue and set (generic implementations)
- **Cache** (`utils/cache`): generic LRU cache (concurrency-safe, TTL support)
- **Maps** (`utils/maps`): concurrent map implementations

### Database Utilities

- **DB** (`utils/db`): database backup tools — mysqldump wrapper (password passed via env var)

### File Operations

- **Files** (`utils/files`): file manipulation utilities
- **Tar** (`utils/tar`): tar archive extraction (with Zip Slip protection)
- **Zip** (`utils/zip`): zip archive operations

### Data Processing

- **JSON** (`utils/json`): JSON processing utilities
- **Array** (`utils/array`): array manipulation functions
- **String** (`utils/string`): string utility functions
- **Random** (`utils/random`): random string generation (not cryptographically secure)

### Network & Web

- **Net** (`utils/net`): network utilities
- **UA** (`utils/ua`): user-agent parsing utilities
- **Resty** (`utils/resty`): resty request middleware (CAT transactions, retry conditions)

### Security & Authentication

- **JWT** (`utils/jwt`): JWT token generation and validation (HS256 only)
- **Crypto** (`utils/crypto`): AES-256-CBC encryption and PKCS5 padding

### Time & Context

- **Time** (`utils/time`): time utility functions
- **Trace** (`utils/trace`): trace id context utilities
- **Proto** (`utils/proto`): protobuf utilities
- **Shell** (`utils/shell`): shell command execution

### Other Utilities

- **Base** (`utils/base`): base utility functions
- **Vars** (`utils/vars`): variable utilities
- **Struct** (`utils/struct.go`): struct reflection utilities

## Static

The `static` directory contains configuration constants, environment variable helpers, and error definitions:

- **Cons** (`static/cons`): application constants
- **Env** (`static/env`): environment variable helpers
- **Errors** (`static/errors`): error definitions

## Model

The `model` directory contains shared data models (monitoring stats, OSS config, etc.).

## Analysis (custom vet)

The `analysis` package provides custom go/analysis-based static checks (e.g. requiring `redis.Generator.GenKey` arguments to be constants from the cons package). `analysis/cmd` is the standalone entry point.

## Test Helpers

The `test` package provides test suite helpers (`BaseSuite`) and the integration test switch (`GO_TOOLS_TEST_INTEGRATION=1` enables tests requiring external environments).

## Usage Examples

### Kafka Producer

```go
import (
	"github.com/alomerry/go-tools/components/kafka"
	"github.com/segmentio/kafka-go"
)

producer, err := kafka.NewDefaultProducer(ctx,
	kafka.WithTopic("my-topic"),
	kafka.WithAddress("localhost:9092"),
)
if err != nil {
	log.Fatal(err)
}
defer producer.Close()

err = producer.Write(ctx, kafka.Message{
	Key:   []byte("key"),
	Value: []byte(`{"value": 1}`),
})
```

### System Monitor

```go
import "github.com/alomerry/go-tools/components/monitor"

m := monitor.NewSystemMonitor(
	monitor.WithContext(ctx),
	monitor.WithInterval(30*time.Second),
	monitor.WithCallback(func(stats *monitor.SystemStats) error {
		log.Printf("CPU: %.2f%%, Memory: %.2f%%",
			stats.CpuUsage, stats.MemoryUsage)
		return nil
	}),
)

m.Watch()
```

### MySQL Dump

```go
import (
	"github.com/alomerry/go-tools/static/cons"
	mysqlDump "github.com/alomerry/go-tools/utils/db/mysql"
)

var tool mysqlDump.DumpTool
path, err := tool.Dump("/tmp/backups", map[string]any{
	"user":     "root",
	"host":     "localhost",
	"port":     "3306",
	"password": os.Getenv("MYSQL_PWD"),
}, cons.Database{Name: "mydb"})
```

### Set Operations

```go
import "github.com/alomerry/go-tools/utils/algorithm"

set := algorithm.Instance[string]()
set.Insert("a")
set.Insert("b")

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
