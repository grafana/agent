# ctimefmt
[![GoDoc](https://godoc.org/github.com/Mottl/ctimefmt?status.svg)](https://godoc.org/github.com/Mottl/ctimefmt)

strptime/strftime compatible syntax (e.g. `"%Y-%m-%d %H:%M:%S %Z"`) for Go.


## Installation
```sh
go get github.com/bluemedora/ctimefmt
```

## Example
```go
package main

import (
    "fmt"
    "time"
    "github.com/Mottl/ctimefmt"
)

func main() {
    now := time.Now()
    // Format() function formats Time struct:
    fmt.Println(ctimefmt.Format("%Y-%m-%d %H:%M:%S.%f %Z", now))

    // ToNative() converts ctime-like format string to Go native layout:
    s := ctimefmt.ToNative("%Y-%m-%d %H:%M:%S.%f %Z")
    fmt.Println(now.Format(s)))

    // Parse() parses ctime-like syntax to Time struct:
    if then, err := ctimefmt.Parse("%Y-%m-%d %H:%M:%S", "2019-02-19 17:25:05"); err == nil {
        fmt.Println(then)
    } else {
        fmt.Println("Error parsing time:", err)
    }
}
```

## License
Use of this package is governed by MIT license
that can be found in the LICENSE file.
