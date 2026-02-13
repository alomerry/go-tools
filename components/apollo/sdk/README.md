# Apollo Open API SDK for Go

This is a Go SDK for Apollo Open API.

## Installation

```bash
go get github.com/alomerry/go-tools/components/apollo/sdk
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/alomerry/go-tools/components/apollo/sdk"
)

func main() {
    client := sdk.NewClient("http://localhost:8070", "your-token")

    // Get Apps
    apps, err := client.Apps.GetApps(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }
    for _, app := range apps {
        fmt.Printf("App: %s\n", app.Name)
    }

    // Get Namespace
    ns, err := client.Namespaces.Get(context.Background(), "DEV", "appId", "default", "application")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Namespace: %+v\n", ns)
}
```

## Features

- App Management
- Cluster Management
- Namespace Management
- Item Management
- Release Management
