# vertica-sql-go

[![License](https://img.shields.io/badge/License-Apache%202.0-orange.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/vertica/vertica-sql-go.svg)](https://pkg.go.dev/github.com/vertica/vertica-sql-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/vertica/vertica-sql-go)](https://goreportcard.com/report/github.com/vertica/vertica-sql-go)

vertica-sql-go is a native Go adapter for the Vertica (http://www.vertica.com) database.

Please check out [release notes](https://github.com/vertica/vertica-sql-go/releases) to learn about the latest improvements.

vertica-sql-go has been tested with Vertica 12.0.0 and Go 1.14/1.15/1.16/1.17/1.18.

## Installation

Source code for vertica-sql-go can be found at:

https://github.com/vertica/vertica-sql-go

Alternatively you can use the 'go get' variant to install the package into your local Go environment.

```sh
go get github.com/vertica/vertica-sql-go
```

## Usage

As this library is written to Go's SQL standard [database/sql](https://golang.org/pkg/database/sql/), usage is compliant with its methods and behavioral expectations.

### Importing

First ensure that you have the library checked out in your standard Go hierarchy and import it.

```Go
import (
    "context"
    "database/sql"
    "github.com/vertica/vertica-sql-go"
)
```

### Setting the Log Level

The vertica-sql-go driver supports multiple log levels, as defined in the following table

| Log Level (int) | Log Level Name | Description |
|-----------------|----------------|-------------|
| 0               | TRACE          | Show function calls, plus all below |
| 1               | DEBUG          | Show low-level functional operations, plus all below |
| 2               | INFO           | Show important state information, plus all below |
| 3               | WARN           | (default) Show non-breaking abnormalities, plus all below |
| 4               | ERROR          | Show breaking errors, plus all below |
| 5               | FATAL          | Show process-breaking errors |
| 6               | NONE           | Disable all log messages |

and they can be set programmatically by calling the logger global level itself

```Go
logger.SetLogLevel(logger.DEBUG)
```

or by setting the environment variable VERTICA_SQL_GO_LOG_LEVEL to one of the integer values in the table above. This must be done before the process using the driver has started as the global log level will be read from here on start-up.

Example:

```bash
export VERTICA_SQL_GO_LOG_LEVEL=3
```

### Setting the Log File

By default, log messages are sent to stdout, but the vertica-sql-go driver can also output to a file in cases where stdout is not available.
Simply set the environment variable VERTICA_SQL_GO_LOG_FILE to your desired output location.

Example:

```bash
export VERTICA_SQL_GO_LOG_FILE=/var/log/vertica-sql-go.log
```

### Creating a connection

```Go
connDB, err := sql.Open("vertica", myDBConnectString)
```

where *myDBConnectString* is of the form:

```Go
vertica://(user):(password)@(host):(port)/(database)?(queryArgs)
```

Currently supported query arguments are:

| Query Argument | Description | Values |
|----------------|-------------|--------|
| use_prepared_statements    | whether to use client-side query interpolation or server-side argument binding | 1 = (default) use server-side bindings |
|                |             | 0 = user client side interpolation **(LESS SECURE)** |
| connection_load_balance    | whether to enable connection load balancing on the client side | 0 = (default) disable load balancing |
|                |             | 1 = enable load balancing |
| tlsmode            | the ssl/tls policy for this connection | 'none' (default) = don't use SSL/TLS for this connection |
|                |                                    | 'server' = server must support SSL/TLS, but skip verification **(INSECURE!)** |
|                |                                    | 'server-strict' = server must support SSL/TLS |
|                |                                    | {customName} = use custom registered `tls.Config` (see "Using custom TLS config" section below) |
| backup_server_node    | a list of backup hosts for the client to try to connect if the primary host is unreachable | a comma-seperated list of backup host-port pairs. E.g.<br> 'host1:port1,host2:port2,host3:port3'  |
| client_label   | Sets a label for the connection on the server. This value appears in the `client_label` column of the SESSIONS system table. | (default) vertica-sql-go-{version}-{pid}-{timestamp} |

To ping the server and validate a connection (as the connection isn't necessarily created at that moment), simply call the *PingContext()* method.

```Go
ctx := context.Background()

err = connDB.PingContext(ctx)
```

If there is an error in connection, the error result will be non-nil and contain a description of whatever problem occurred.

### Using custom TLS config

Custom TLS config(s) can be registered for TLS / SSL encrypted connection to the server.
Here is an example of registering and using a `tls.Config`:

```Go
import vertigo "github.com/vertica/vertica-sql-go"

// Register tls.Config
rootCertPool := x509.NewCertPool()
pem, err := ioutil.ReadFile("/certs/ca.crt")
if err != nil {
    LOG.Warningln("ERROR: failed reading cert file", err)
}
if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
    LOG.Warningln("ERROR: Failed to append PEM")
}
tlsConfig := &tls.Config{RootCAs: rootCertPool, ServerName: host}
vertigo.RegisterTLSConfig("myCustomName", tlsConfig)

// Connect using tls.Config
var rawQuery = url.Values{}
rawQuery.Add("tlsmode", "myCustomName")
var query = url.URL{
    Scheme:   "vertica",
    User:     url.UserPassword(user, password),
    Host:     fmt.Sprintf("%s:%d", host, port),
    Path:     databaseName,
    RawQuery: rawQuery.Encode(),
}
sql.Open("vertica", query.String())
```

### Performing a simple query

Performing a simple query is merely a matter of using that connection to create a query and iterate its results.
Here is an example of a query that should always work.

```Go
rows, err := connDB.QueryContext(ctx, "SELECT * FROM v_monitor.cpu_usage LIMIT 5")

defer rows.Close()
```

**IMPORTANT** : Just as with connections, you should always Close() the results cursor once you are done with it. It's often easier to just defer the closure, for convenience.

### Performing a query with arguments

This is done in a similar manner on the client side.

```Go
rows, err := connDB.QueryContext(ctx, "SELECT name FROM MyTable WHERE id=?", 21)
```

Behind the scenes, this will be handled in one of two ways, based on whether or not you requested client interpolation in the connection string.

With client interpolation enabled, the client library will create a new query string with the arguments already in place, and submit it as a simple query.

With client interpolation disabled (default), the client library will use the full server-side parse(), describe(), bind(), execute() cycle.

#### Named Arguments

```Go
rows, err := connDB.QueryContext(ctx, "SELECT name FROM MyTable WHERE id=@id and something=@example", sql.Named("id", 21), sql.Named("example", "hello"))
```

Named arguments are emulated by the driver. They will be converted to positional arguments by the driver and the named arguments given later will be slotted
into the required positions. This still allows server side prepared statements as `@id` and `@example` above will be replaced by `?` before being sent. If
you use named arguments, all the arguments must be named. Do not mix positional and named together. All named arguments are normalized to upper case which means
`@param`, `@PaRaM`, and `@PARAM` are treated as equivalent.

### Reading query result rows

As outlined in the GoLang specs, reading the results of a query is done via a loop, bounded by a .next() iterator.

```Go
for rows.Next() {
    var nodeName string
    var startTime string
    var endTime string
    var avgCPU float64

    rows.Scan(&nodeName, &startTime, &endTime, &avgCPU)

    // Use these values for something here.
}
```

If you need to examine the names of the columns, simply access the Columns() operator of the rows object.

```Go
columnNames, _ := rows.Columns()

for _, columnName := range columnNames {
        // use the column name here.
}
```

### Paging in Data

By default, the query results are cached in memory allowing for rapid iteration of result row content.
This generally works well, but in the case of exceptionally large result sets, you could run out of memory.

If such a query needs to be performed, it is recommended that you tell the driver that you wish to cache
that data in a temporary file, so its results can be "paged in" as you iterate the results. The data is
stored in a process-read-only file in the OS's temp directory.

To enable result paging, simply create a VerticaContext and use it to perform your query.

```go
vCtx := NewVerticaContext(context.Background())

// Only keep 50000 rows in memory at once.
vCtx.SetInMemoryResultRowLimit(50000)

rows, _ := connDB.QueryContext(
    vCtx,
    "SELECT a, b, c, d, e FROM result_cache_test ORDER BY a")

defer rows.Close()

// Use rows result as normal.
```

If you want to disable paging on the same context all together, you can simply set the row
limit to 0 (the default).

### Performing a simple execute call

This is very similar to a simple query, but has a slightly different result type. A simple execute() might look like this:

```Go
res, err = connDB.ExecContext(ctx, "DROP TABLE IF EXISTS MyTable")
```

In this instance, *res* will contain information (such as 'rows affected') about the result of this execution.

### Performing an execute with arguments

This, again, looks very similar to the query-with-arguments use case and is subject to the same effects of client-side interpolation.

```Go
res, err := connDB.ExecContext(
        ctx,
        "INSERT INTO MyTable VALUES (?)", 21)
```

### Server-side prepared statements

**IMPORTANT** : Vertica does not support executing a command string containing multiple statements using server-side prepared statements.

If you wish to reuse queries or executions, you can prepare them once and supply arguments only.

```Go
// Prepare the query.
stmt, err := connDB.PrepareContext(ctx, "SELECT id FROM MyTable WHERE name=?")

// Execute it with this argument.
rows, err = stmt.Query("Joe Perry")
```

**NOTE** : Please note that this method is subject to modification by the 'interpolate' setting. If the client side interpolation is requested, the statement will simply be stored on the client and interpolated with arguments each time it's used. If not using client side interpolation (default), the statement will be parsed and described on the server as expected.

### Transactions

The vertica-sql-go driver supports basic transactions as defined by the GoLang standard.

```Go
// Define the options for this transaction state
opts := &sql.TxOptions{
    Isolation: sql.LevelDefault,
    ReadOnly:  false,
}

// Begin the transaction.
tx, err := connDB.BeginTx(ctx, opts)
```

```Go
// You can either commit it.
err = tx.Commit()
```

```Go
// Or roll it back.
err = tx.Rollback()
```

The following transaction isolation levels are supported:

* sql.LevelReadUncommitted <sup><b>&#8224;</b></sup>
* sql.LevelReadCommitted
* sql.LevelSerializable
* sql.LevelRepeatableRead <sup><b>&#8224;</b></sup>
* sql.LevelDefault

 The following transaction isolation levels are unsupported:

* sql.LevelSnapshot
* sql.LevelLinearizable

 <b>&#8224;</b> Although Vertica supports the grammars for these transaction isolation levels, they are internally promoted to stronger isolation levels.

## COPY modes Supported

### COPY FROM STDIN

vertica-sql-go supports copying from stdin. This allows you to write a command-line tool that accepts stdin as an
input and passes it to Vertica for processing. An example:

```go
_, err = connDB.ExecContext(ctx, "COPY stdin_data FROM STDIN DELIMITER ','")
```

This will process input from stdin until an EOF is reached.

### COPY FROM STDIN with alternate stream

In your code, you may also supply a different io.Reader object (such as *File) from which to supply your data.
Simply create a new VerticaContext, set the copy input stream, and provide this context to the execute call.
An example:

```go
fp, err := os.OpenFile("./resources/csv/sample_data.csv", os.O_RDONLY, 0600)
...
vCtx := NewVerticaContext(ctx)
vCtx.SetCopyInputStream(fp)

_, err = connDB.ExecContext(vCtx, "COPY stdin_data FROM STDIN DELIMITER ','")
```

If you provide a VerticaContext but don't set a copy input stream, the driver will fall back to os.stdin.

## Full Example

By following the above instructions, you should be able to successfully create a connection to your Vertica instance and perform the operations you require. A complete example program is listed below:

```Go
package main

import (
    "context"
    "database/sql"
    "os"

    _ "github.com/vertica/vertica-sql-go"
    "github.com/vertica/vertica-sql-go/logger"
)

func main() {
    // Have our logger output INFO and above.
    logger.SetLogLevel(logger.INFO)

    var testLogger = logger.New("samplecode")

    ctx := context.Background()

    // Create a connection to our database. Connection is lazy and won't
    // happen until it's used.
    connDB, err := sql.Open("vertica", "vertica://dbadmin:@localhost:5433/db1?connection_load_balance=1")

    if err != nil {
        testLogger.Fatal(err.Error())
        os.Exit(1)
    }

    defer connDB.Close()

    // Ping the database connnection to force it to attempt to connect.
    if err = connDB.PingContext(ctx); err != nil {
        testLogger.Fatal(err.Error())
        os.Exit(1)
    }

    // Query a standard metric table in Vertica.
    rows, err := connDB.QueryContext(ctx, "SELECT * FROM v_monitor.cpu_usage LIMIT 5")

    if err != nil {
        testLogger.Fatal(err.Error())
        os.Exit(1)
    }

    defer rows.Close()

    // Iterate over the results and print them out.
    for rows.Next() {
        var nodeName string
        var startTime string
        var endTime string
        var avgCPU float64

        if err = rows.Scan(&nodeName, &startTime, &endTime, &avgCPU); err != nil {
            testLogger.Fatal(err.Error())
            os.Exit(1)
        }

        testLogger.Info("%s\t%s\t%s\t%f", nodeName, startTime, endTime, avgCPU)
    }

    testLogger.Info("Test complete")

    os.Exit(0)
}
```

## License

Apache 2.0 License, please see `LICENSE` for details.

## Contributing guidelines

Have a bug or an idea? Please see `CONTRIBUTING.md` for details.

### Benchmarks

You can run a benchmark and profile it with a command like:
`go test -bench '^BenchmarkRowsWithLimit$' -benchmem -memprofile memprofile.out -cpuprofile profile.out -run=none`

and then explore it with `go tool pprof`. The `-run` part excludes the tests for brevity.

## Acknowledgements
* @grzm (Github)
* @watercraft (Github)
* @fbernier (Github)
* @mlh758 (Github) for the awesome work filling in and enhancing the driver in many important ways.
* Tom Wall (Vertica) for the infinite patience and deep knowledge.
* The creators and contributors of the vertica-python library, and members of the Vertica team, for their help in understanding the wire protocol.
