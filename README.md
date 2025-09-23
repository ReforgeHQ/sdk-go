# prefab-cloud-go

Go Client for Prefab Feature Flags, Dynamic log levels, and Config as a Service: https://www.prefab.cloud

## Installation

```bash
go get github.com/prefab-cloud/prefab-cloud-go@latest
```

## Basic example

```go
package main

import (
	"fmt"
	"log"
	"os"

	reforge "github.com/ReforgeHQ/sdk-go/pkg"
)

func main() {
	sdkKey, exists := os.LookupEnv("REFORGE_SDK_KEY")

	if !exists {
		log.Fatal("SDK Key not found")
	}

	client, err := reforge.NewSdk(reforge.WithSdkKey(sdkKey))

	if err != nil {
		log.Fatal(err)
	}

	val, ok, err := client.GetStringValue("my.string.config", reforge.ContextSet{})

	if err != nil {
		log.Fatal(err)
	}

	if !ok {
		log.Fatal("Value not found")
	}

	fmt.Println(val)
}
```

## Documentation

- [API Reference](https://pkg.go.dev/github.com/prefab-cloud/prefab-cloud-go/pkg)

## Notable pending features

- Telemetry


## Publishing 

1) Bump version in pkg/internal/version.go (this is the version header clients send)
2) Commit that change on a branch and merge into main
3) git tag with the new version number and push that to origin 
