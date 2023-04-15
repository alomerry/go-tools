# go-pusher

go-pusher 是一个可以帮助您上传文件到 OSS 并定期将其备份到 VPS 上的工具

## 要求

- `Go 1.20` 及以上

## 功能

- [x] 上传指定目录下 OSS 中不存在或和 OSS 中不一致的文件
- [ ] 定时备份本地文件到 VPS 上

## 使用

- 构建二进制文件

  `go build main.go`

- 添加可执行权限

  `chmod +x ./main`

- 通过配置文件运行

  `./main -configPath "配置文件绝对路径"`

## 配置文件

以下是配置的内容：

```toml
modes = ["pusher", "syncer"]

[syncer]
# 本地绝对路径
local-path = "xxx"
# 绝对路径
remote-path = "xxx"
# 检查文件变动间隔（秒）
interval = 1

[pusher]
# oss
oss-provider = "qiniu"
# 待推送目录对应 oss 的前缀
oss-object-prefix = "blog/public"
# 推送到 oss 超时时间（秒）
push-timeout = 60
# 本地待检测文件夹绝对路径
local-directory = "/path/to/push"
# 是否和本地强一致，即本地不存在的文件会从 oss 中删除
oss-delete-not-exists = false

[oss-qiniu]
bucket = "xxx"
# 存储区域
# 华东 z0、华东浙江 2 区 cn-east-2、华北 z1、华南 z2、北美 na0、新加坡 as0、亚太首尔 1 区 fog-cn-east-1（详见 https://developer.qiniu.com/kodo/1238/go）
region = "ZoneHuadong"
access-key = "xxx"
sercet-key = "xxx"
```

## 为什么写了 go-pusher

## [LICENSE](LICENSE)
