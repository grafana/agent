# About dburl

Package `dburl` provides a standard, URL style mechanism for parsing and
opening SQL database connection strings for [Go][go-project]. Provides
standardized way to [parse][goref-parse] and [open][goref-open] URLs for
popular databases PostgreSQL, MySQL, SQLite3, Oracle Database, Microsoft SQL
Server, in addition to most other SQL databases with a publicly available Go
driver.

[Overview][] | [Quickstart][] | [Examples][] | [Schemes][] | [Installing][] | [Using][] | [About][]

[Overview]: #database-connection-url-overview (Database Connection URL Overview)
[Quickstart]: #quickstart (Quickstart)
[Examples]: #example-urls (Example URLs)
[Schemes]: #protocol-schemes-and-aliases (Protocol Schemes and Aliases)
[Installing]: #installing (Installing)
[Using]: #using (Using)
[About]: #about (About)

[![Unit Tests][dburl-ci-status]][dburl-ci]
[![Go Reference][goref-dburl-status]][goref-dburl]
[![Discord Discussion][discord-status]][discord]

[dburl-ci]: https://github.com/xo/dburl/actions/workflows/test.yml
[dburl-ci-status]: https://github.com/xo/dburl/actions/workflows/test.yml/badge.svg
[goref-dburl]: https://pkg.go.dev/github.com/xo/dburl
[goref-dburl-status]: https://pkg.go.dev/badge/github.com/xo/dburl.svg
[discord]: https://discord.gg/yJKEzc7prt (Discord Discussion)
[discord-status]: https://img.shields.io/discord/829150509658013727.svg?label=Discord&logo=Discord&colorB=7289da&style=flat-square (Discord Discussion)

## Database Connection URL Overview

Supported database connection URLs are of the form:

```text
protocol+transport://user:pass@host/dbname?opt1=a&opt2=b
protocol:/path/to/file
```

Where:

| Component          | Description                                                                          |
|--------------------|--------------------------------------------------------------------------------------|
| protocol           | driver name or alias (see below)                                                     |
| transport          | "tcp", "udp", "unix" or driver name (odbc/oleodbc)                                   |
| user               | username                                                                             |
| pass               | password                                                                             |
| host               | host                                                                                 |
| dbname<sup>*</sup> | database, instance, or service name/ID to connect to                                 |
| ?opt1=...          | additional database driver options (see respective SQL driver for available options) |

<i><sup><b>*</b></sup> for Microsoft SQL Server, `/dbname` can be
`/instance/dbname`, where `/instance` is optional. For Oracle Database,
`/dbname` is of the form `/service/dbname` where `/service` is the service name
or SID, and `/dbname` is optional. Please see below for examples.</i>

## Quickstart

Database connection URLs in the above format can be parsed with the
[`dburl.Parse` func][goref-parse] as such:

```go
import (
    "github.com/xo/dburl"
)
u, err := dburl.Parse("postgresql://user:pass@localhost/mydatabase/?sslmode=disable")
if err != nil { /* ... */ }
```

Additionally, a simple helper, [`dburl.Open`][goref-open], is provided that
will parse, open, and return a [standard `sql.DB` database][goref-sql-db]
connection:

```go
import (
    "github.com/xo/dburl"
)
db, err := dburl.Open("sqlite:mydatabase.sqlite3?loc=auto")
if err != nil { /* ... */ }
```

## Example URLs

The following are example database connection URLs that can be handled by
[`dburl.Parse`][goref-parse] and [`dburl.Open`][goref-open]:

```text
postgres://user:pass@localhost/dbname
pg://user:pass@localhost/dbname?sslmode=disable
mysql://user:pass@localhost/dbname
mysql:/var/run/mysqld/mysqld.sock
sqlserver://user:pass@remote-host.com/dbname
mssql://user:pass@remote-host.com/instance/dbname
ms://user:pass@remote-host.com:port/instance/dbname?keepAlive=10
oracle://user:pass@somehost.com/sid
sap://user:pass@localhost/dbname
sqlite:/path/to/file.db
file:myfile.sqlite3?loc=auto
odbc+postgres://user:pass@localhost:port/dbname?option1=
```

## Protocol Schemes and Aliases

The following protocols schemes (ie, driver) and their associated aliases are
supported out of the box:

<!-- START SCHEME TABLE -->
| Database (scheme/driver)         | Protocol Aliases [real driver]           |
|----------------------------------|------------------------------------------|
| MySQL (mysql)                    | my, mariadb, maria, percona, aurora      |
| Oracle Database (oracle)         | or, ora, oracle, oci, oci8, odpi, odpi-c |
| PostgreSQL (postgres)            | pg, postgresql, pgsql                    |
| SQLite3 (sqlite3)                | sq, sqlite, file                         |
| Microsoft SQL Server (sqlserver) | ms, mssql, azuresql                      |
|                                  |                                          |
| Amazon Redshift (redshift)       | rs [postgres]                            |
| CockroachDB (cockroachdb)        | cr, cockroach, crdb, cdb [postgres]      |
| MemSQL (memsql)                  | me [mysql]                               |
| TiDB (tidb)                      | ti [mysql]                               |
| Vitess (vitess)                  | vt [mysql]                               |
|                                  |                                          |
| MySQL (mymysql)                  | zm, mymy                                 |
| Oracle Database (godror)         | gr                                       |
| PostgreSQL (pgx)                 | px                                       |
|                                  |                                          |
| Alibaba MaxCompute (maxcompute)  | mc                                       |
| Alibaba Tablestore (ots)         | ot, ots, tablestore                      |
| Apache Avatica (avatica)         | av, phoenix                              |
| Apache H2 (h2)                   | h2                                       |
| Apache Hive (hive)               | hi                                       |
| Apache Ignite (ignite)           | ig, gridgain                             |
| Apache Impala (impala)           | im                                       |
| AWS Athena (awsathena)           | s3, aws, athena                          |
| Azure Cosmos (cosmos)            | cm                                       |
| Cassandra (cql)                  | ca, cassandra, datastax, scy, scylla     |
| ClickHouse (clickhouse)          | ch                                       |
| Couchbase (n1ql)                 | n1, couchbase                            |
| Cznic QL (ql)                    | ql, cznic, cznicql                       |
| CSVQ (csvq)                      | csv, tsv, json                           |
| Databend (databend)              | dd, bend                                 |
| Exasol (exasol)                  | ex, exa                                  |
| Firebird SQL (firebirdsql)       | fb, firebird                             |
| Genji (genji)                    | gj                                       |
| Google BigQuery (bigquery)       | bq                                       |
| Google Spanner (spanner)         | sp                                       |
| IBM Netezza (nzgo)               | nz, netezza                              |
| Microsoft ADODB (adodb)          | ad, ado                                  |
| ModernC SQLite (moderncsqlite)   | mq, modernsqlite                         |
| ODBC (odbc)                      | od                                       |
| OLE ODBC (oleodbc)               | oo, ole, oleodbc [adodb]                 |
| Presto (presto)                  | pr, prestodb, prestos, prs, prestodbs    |
| SAP ASE (tds)                    | ax, ase, sapase                          |
| SAP HANA (hdb)                   | sa, saphana, sap, hana                   |
| Snowflake (snowflake)            | sf                                       |
| Trino (trino)                    | tr, trino, trinos, trs                   |
| Vertica (vertica)                | ve                                       |
| VoltDB (voltdb)                  | vo, volt, vdb                            |
<!-- END SCHEME TABLE -->

Any protocol scheme `alias://` can be used in place of `protocol://`, and will
work identically with [`dburl.Parse`][goref-parse] and [`dburl.Open`][goref-open].

## Installing

Install in the usual Go fashion:

```sh
$ go get -u github.com/xo/dburl
```

## Using

Please note that `dburl` does not import actual SQL drivers, and only provides
a standard way to [parse][goref-parse]/[open][goref-open] respective database
connection URLs.

For reference, these are the following "expected" SQL drivers that would need
to be imported:

<!-- START DRIVER TABLE -->
| Database (driver)                | Package                                                                                                        |
|----------------------------------|----------------------------------------------------------------------------------------------------------------|
| MySQL (mysql)                    | [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)                                       |
| Oracle Database (oracle)         | [github.com/sijms/go-ora](https://github.com/sijms/go-ora)                                                     |
| PostgreSQL (postgres)            | [github.com/lib/pq](https://github.com/lib/pq)                                                                 |
| SQLite3 (sqlite3)                | [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)                                             |
| Microsoft SQL Server (sqlserver) | [github.com/microsoft/go-mssqldb](https://github.com/microsoft/go-mssqldb)                                     |
|                                  |                                                                                                                |
| Amazon Redshift (redshift)       | [github.com/lib/pq](https://github.com/lib/pq)                                                                 |
| CockroachDB (cockroachdb)        | [github.com/lib/pq](https://github.com/lib/pq)                                                                 |
| MemSQL (memsql)                  | [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)                                       |
| TiDB (tidb)                      | [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)                                       |
| Vitess (vitess)                  | [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)                                       |
|                                  |                                                                                                                |
| MySQL (mymysql)                  | [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql)                                           |
| Oracle Database (godror)         | [github.com/godror/godror](github.com/godror/godror)                                                           |
| PostgreSQL (pgx)                 | [github.com/jackc/pgx/stdlib](https://github.com/jackc/pgx)                                                    |
|                                  |                                                                                                                |
| Alibaba MaxCompute (maxcompute)  | [sqlflow.org/gomaxcompute](https://sqlflow.org/gomaxcompute)                                                   |
| Alibaba Tablestore (ots)         | [github.com/aliyun/aliyun-tablestore-go-sql-driver](https://github.com/aliyun/aliyun-tablestore-go-sql-driver) |
| Apache Avatica (avatica)         | [github.com/Boostport/avatica](https://github.com/Boostport/avatica)                                           |
| Apache H2 (h2)                   | [github.com/jmrobles/h2go](https://github.com/jmrobles/h2go)                                                   |
| Apache Hive (hive)               | [sqlflow.org/gohive](https://sqlflow.org/gohive)                                                               |
| Apache Ignite (ignite)           | [github.com/amsokol/ignite-go-client/sql](https://github.com/amsokol/ignite-go-client)                         |
| Apache Impala (impala)           | [github.com/bippio/go-impala](https://github.com/bippio/go-impala)                                             |
| AWS Athena (awsathena)           | [github.com/uber/athenadriver/go](https://github.com/uber/athenadriver)                                        |
| Azure Cosmos (cosmos)            | [github.com/btnguyen2k/gocosmos](https://github.com/btnguyen2k/gocosmos)                                       |
| Cassandra (cql)                  | [github.com/MichaelS11/go-cql-driver](https://github.com/MichaelS11/go-cql-driver)                             |
| ClickHouse (clickhouse)          | [github.com/ClickHouse/clickhouse-go](https://github.com/ClickHouse/clickhouse-go)                             |
| Couchbase (n1ql)                 | [github.com/couchbase/go_n1ql](https://github.com/couchbase/go_n1ql)                                           |
| Cznic QL (ql)                    | [modernc.org/ql](https://modernc.org/ql)                                                                       |
| CSVQ (csvq)                      | [github.com/mithrandie/csvq](https://github.com/mithrandie/csvq)                                               |
| Databend (databend)              | [github.com/databendcloud/databend-go](https://github.com/databendcloud/databend-go)                           |
| Exasol (exasol)                  | [github.com/exasol/exasol-driver-go](https://github.com/exasol/exasol-driver-go)                               |
| Firebird SQL (firebirdsql)       | [github.com/nakagami/firebirdsql](https://github.com/nakagami/firebirdsql)                                     |
| Genji (genji)                    | [github.com/genjidb/genji/sql/driver](https://github.com/genjidb/genji)                                        |
| Google BigQuery (bigquery)       | [gorm.io/driver/bigquery/driver](https://gorm.io/driver/bigquery/driver)                                       |
| Google Spanner (spanner)         | [github.com/rakyll/go-sql-driver-spanner](https://github.com/rakyll/go-sql-driver-spanner)                     |
| IBM Netezza (nzgo)               | [github.com/IBM/nzgo](https://github.com/IBM/nzgo)                                                             |
| Microsoft ADODB (adodb)          | [github.com/mattn/go-adodb](https://github.com/mattn/go-adodb)                                                 |
| ModernC SQLite (moderncsqlite)   | [modernc.org/sqlite](https://modernc.org/sqlite)                                                               |
| ODBC (odbc)                      | [github.com/alexbrainman/odbc](https://github.com/alexbrainman/odbc)                                           |
| OLE ODBC (oleodbc)               | [github.com/mattn/go-adodb](https://github.com/mattn/go-adodb)                                                 |
| Presto (presto)                  | [github.com/prestodb/presto-go-client/presto](https://github.com/prestodb/presto-go-client)                    |
| SAP ASE (tds)                    | [github.com/thda/tds](https://github.com/thda/tds)                                                             |
| SAP HANA (hdb)                   | [github.com/SAP/go-hdb/driver](https://github.com/SAP/go-hdb)                                                  |
| Snowflake (snowflake)            | [github.com/snowflakedb/gosnowflake](https://github.com/snowflakedb/gosnowflake)                               |
| Trino (trino)                    | [github.com/trinodb/trino-go-client/trino](https://github.com/trinodb/trino-go-client)                         |
| Vertica (vertica)                | [github.com/vertica/vertica-sql-go](https://github.com/vertica/vertica-sql-go)                                 |
| VoltDB (voltdb)                  | [github.com/VoltDB/voltdb-client-go/voltdbclient](github.com/VoltDB/voltdb-client-go])                         |
<!-- END DRIVER TABLE -->

Please see [the `dburl` Go Reference][goref-link] for the full API
documentation.

### URL Parsing Rules

[`dburl.Parse`][goref-parse] and [`dburl.Open`][goref-open] rely primarily on
Go's standard [`net/url.URL`][goref-net-url] type, and as such, parsing or
opening database connection URLs with `dburl` are subject to the same rules,
conventions, and semantics as [Go's `net/url.Parse` func][goref-net-url-parse].

## Example

A [full example](_example/example.go) for reference:

```go
// _example/example.go
package main

import (
	"fmt"
	"log"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/xo/dburl"
)

func main() {
	db, err := dburl.Open("sqlserver://user:pass@localhost/dbname")
	if err != nil {
		log.Fatal(err)
	}
	var name string
	if err := db.QueryRow(`SELECT name FROM mytable WHERE id=10`).Scan(&name); err != nil {
		log.Fatal(err)
	}
	fmt.Println("name:", name)
}
```

## About

`dburl` was built primarily to support these projects:

* [usql][usql] - a universal command-line interface for SQL databases
* [xo][xo] - a command-line tool to generate code for SQL databases

[go-project]: https://golang.org/project
[goref-open]: https://pkg.go.dev/github.com/xo/dburl#Open
[goref-parse]: https://pkg.go.dev/github.com/xo/dburl#Parse
[goref-sql-db]: https://pkg.go.dev/database/sql#DB
[goref-net-url]: https://pkg.go.dev/net/url#URL
[goref-net-url-parse]: https://pkg.go.dev/net/url#URL.Parse

[usql]: https://github.com/xo/usql
[xo]: https://github.com/xo/xo
