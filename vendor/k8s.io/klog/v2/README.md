# klog-gokit [![CircleCI](https://circleci.com/gh/simonpasquier/klog-gokit.svg?style=svg)](https://circleci.com/gh/simonpasquier/klog-gokit)

This package is a replacement for [k8s.io/klog/v2](https://github.com/kubernetes/klog)
in projects that use the [github.com/go-kit/log](https://pkg.go.dev/github.com/go-kit/log) module for logging.

*The current branch supports neither
[k8s.io/klog](https://pkg.go.dev/k8s.io/klog) nor
[github.com/go-kit/kit](https://pkg.go.dev/github.com/go-kit/kit). Please use the `v2.1.0`
version instead.*

It is heavily inspired by the [`github.com/kubermatic/glog-gokit`](https://github.com/kubermatic/glog-gokit) package.

## Usage

Add this line to your `go.mod` file:

```
replace k8s.io/klog/v2 => github.com/simonpasquier/klog-gokit/v3 v3
```

In your `main.go`:
```go
// Import the package like it is original klog
import (
    ...
    "github.com/go-kit/log"
    klog "k8s.io/klog/v2"
    ...
)

// Create go-kit logger in your main.go
logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
logger = log.With(logger, "ts", log.DefaultTimestampUTC)
logger = log.With(logger, "caller", log.DefaultCaller)
logger = level.NewFilter(logger, level.AllowAll())

// Overriding the default klog with our go-kit klog implementation.
// Thus we need to pass it our go-kit logger object.
klog.SetLogger(logger)
```

Setting the klog's logger **MUST** happen at the very beginning of your program
(e.g. before using the other klog functions).

## Function Levels

|     klog     | gokit |
| ------------ | ----- |
| Info         | Debug |
| InfoDepth    | Debug |
| Infof        | Debug |
| Infoln       | Debug |
| InfoS        | Debug |
| InfoSDepth   | Debug |
| Warning      | Warn  |
| WarningDepth | Warn  |
| Warningf     | Warn  |
| Warningln    | Warn  |
| Error        | Error |
| ErrorDepth   | Error |
| Errorf       | Error |
| Errorln      | Error |
| Exit         | Error |
| ExitDepth    | Error |
| Exitf        | Error |
| Exitln       | Error |
| Fatal        | Error |
| FatalDepth   | Error |
| Fatalf       | Error |
| Fatalln      | Error |

This table is rather opinionated and build for use with the Kubernetes' [Go client](https://github.com/kubernetes/client-go).

## Disclaimer

This project doesn't aim at covering the complete `klog` API. That being said, it should work ok for
projects that use `k8s.io/client-go` (like [Prometheus](https://github.com/prometheus/prometheus) for instance).

## License

Apache License 2.0, see [LICENSE](https://github.com/simonpasquier/klog-gokit/blob/master/LICENSE).
