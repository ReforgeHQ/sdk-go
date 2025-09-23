# sdk-go

Go SDK for Reforge Feature Flags and Config as a Service: https://www.reforge.com

## Installation

```bash
go get github.com/ReforgeHQ/sdk-go@latest
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

	val, ok, err := client.GetStringValue("my.string.config", *reforge.NewContextSet())

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

- [API Reference](https://pkg.go.dev/github.com/ReforgeHQ/sdk-go/pkg)

## Notable pending features

- Telemetry


## Publishing 

1) Bump version in pkg/internal/version.go (this is the version header clients send)
2) Commit that change on a branch and merge into main
3) git tag with the new version number and push that to origin 
