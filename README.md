# go-pusher

English | [简体中文](README_ZH.md)

go-pusher is a tool that can help you upload files to oss and backup file to vps regularly.

## Requirements

- `Go 1.20` and above.

## Features

- [x] upload file to OSS if the file does not exist in OSS or the file in OSS is different from the local file
- [ ] backup local file to VPS regularly

## Usage

- build bin

  `go build main.go`

- add execute permission

  `chmod +x ./main`

- run with config

  `./main -configPath "Your config file abstract path"`

## Config file

Config file is like following:

```toml
modes = ["pusher", "syncer"]

[syncer]
# local directory abstract path
local-path = "xxx"
# remote directory abstract path
remote-path = "xxx"
# time to check file change(second)
interval = 1

[pusher]
# oss provider( now support: qiniu)
oss-provider = "qiniu"
oss-object-prefix = "blog/public"
push-timeout = 60
local-directory = "/path/to/push"
oss-delete-not-exists = false

[oss-qiniu]
bucket = "xxx"
region = "ZoneHuadong"
access-key = "xxx"
sercet-key = "xxx"
```

## why cretea go-pusher

## [LICENSE](LICENSE)
