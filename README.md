# go-tools

go-tools contains several tools

[![go report](https://goreportcard.com/badge/github.com/alomerry/go-tools)](https://goreportcard.com/report/github.com/alomerry/go-tools)

## Requirement

- `Go 1.21` and above.

## Build

`CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./output/go-tools main.go &&upx ./output/go-tools`

## Usage

- [DNS](./dns/README.md)
- [pusher](./pusher/README.md)
- [sgs delay](./sgs/README.md)
- [copier](./copier/README.md)
- algorithm

## Thanks for free JetBrains Open Source license

<a href="https://www.jetbrains.com/?from=alomerry/go-tools" target="_blank">
<img src="https://user-images.githubusercontent.com/1787798/69898077-4f4e3d00-138f-11ea-81f9-96fb7c49da89.png" height="100"/></a>
