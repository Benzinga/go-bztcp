# Go BzTCP
This project provides a pure-Go implementation of the Benzinga TCP protocol.

**Service Deprecated**

# Features

  * Tested with Go 1.13+
  * Reasonably performant.
  * No external dependencies.

# Getting Started
To install the library and the example client, run the following:

```sh
go install github.com/Benzinga/go-bztcp/cmd/bztcp
```

To use the example client, use the new `bztcp` binary. By default, it will be installed to `$GOPATH/bin`. If this isn't on your `$PATH`/`%PATH%` you may need to invoke it by specifying the absolute path of the binary.

```sh
bztcp -v -user USER -key KEY
```

If all has gone well, you should begin seeing messages shortly, depending on the time of day.

This program makes use of the Go context library and thus requires Go 1.8. It would be relatively simple to backport this library to use an external implementation of context.

# Usage
The Go library exposes both high-level and low-level functionality for dealing with the Benzinga TCP protocol, but in particular you usually only need to be concerned with two functions: `Dial`, and `Conn.Stream`.

A quick example follows:

```go
package main

import (
	"context"
	"fmt"

	"github.com/Benzinga/go-bztcp/bztcp"
)

func main() {
	conn, err := bztcp.Dial("tcp-v1.benzinga.io:11337", "USER", "KEY")

	if err != nil {
		panic(err)
	}

	err = conn.Stream(context.Background(), func(stream bztcp.StreamData) {
		fmt.Printf("%#v\n", stream)
	})

	if err != nil {
		panic(err)
	}
}
```
