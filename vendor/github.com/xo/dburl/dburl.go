// Package dburl provides a standard, URL style mechanism for parsing and
// opening SQL database connection strings for Go. Provides standardized way to
// parse and open URLs for popular databases PostgreSQL, MySQL, SQLite3, Oracle
// Database, Microsoft SQL Server, in addition to most other SQL databases with
// a publicly available Go driver.
//
// # Database Connection URL Overview
//
// Supported database connection URLs are of the form:
//
//	protocol+transport://user:pass@host/dbname?opt1=a&opt2=b
//	protocol:/path/to/file
//
// Where:
//
//	protocol  - driver name or alias (see below)
//	transport - "tcp", "udp", "unix" or driver name (odbc/oleodbc)                                  |
//	user      - username
//	pass      - password
//	host      - host
//	dbname*   - database, instance, or service name/id to connect to
//	?opt1=... - additional database driver options
//	              (see respective SQL driver for available options)
//
// * for Microsoft SQL Server, /dbname can be /instance/dbname, where /instance
// is optional. For Oracle Database, /dbname is of the form /service/dbname
// where /service is the service name or SID, and /dbname is optional. Please
// see below for examples.
//
// # Quickstart
//
// Database connection URLs in the above format can be parsed with Parse as such:
//
//	import (
//	    "github.com/xo/dburl"
//	)
//	u, err := dburl.Parse("postgresql://user:pass@localhost/mydatabase/?sslmode=disable")
//	if err != nil { /* ... */ }
//
// Additionally, a simple helper, Open, is provided that will parse, open, and
// return a standard sql.DB database connection:
//
//	import (
//	    "github.com/xo/dburl"
//	)
//	db, err := dburl.Open("sqlite:mydatabase.sqlite3?loc=auto")
//	if err != nil { /* ... */ }
//
// # Example URLs
//
// The following are example database connection URLs that can be handled by
// Parse and Open:
//
//	postgres://user:pass@localhost/dbname
//	pg://user:pass@localhost/dbname?sslmode=disable
//	mysql://user:pass@localhost/dbname
//	mysql:/var/run/mysqld/mysqld.sock
//	sqlserver://user:pass@remote-host.com/dbname
//	mssql://user:pass@remote-host.com/instance/dbname
//	ms://user:pass@remote-host.com:port/instance/dbname?keepAlive=10
//	oracle://user:pass@somehost.com/sid
//	sap://user:pass@localhost/dbname
//	sqlite:/path/to/file.db
//	file:myfile.sqlite3?loc=auto
//	odbc+postgres://user:pass@localhost:port/dbname?option1=
//
// # Protocol Schemes and Aliases
//
// The following protocols schemes (ie, driver) and their associated aliases
// are supported out of the box:
//
//	  Database (scheme/driver)         | Protocol Aliases [real driver]
//	  ---------------------------------|--------------------------------------------
//	  Microsoft SQL Server (sqlserver) | ms, mssql, azuresql
//	  MySQL (mysql)                    | my, mariadb, maria, percona, aurora
//	  Oracle Database (oracle)         | or, ora, oci, oci8, odpi, odpi-c
//	  PostgreSQL (postgres)            | pg, postgresql, pgsql
//	  SQLite3 (sqlite3)                | sq, sqlite, file
//	  ---------------------------------|--------------------------------------------
//	  Amazon Redshift (redshift)       | rs [postgres]
//	  CockroachDB (cockroachdb)        | cr, cockroach, crdb, cdb [postgres]
//	  MemSQL (memsql)                  | me [mysql]
//	  TiDB (tidb)                      | ti [mysql]
//	  Vitess (vitess)                  | vt [mysql]
//	  ---------------------------------|--------------------------------------------
//	  MySQL (mymysql)                  | zm, mymy
//	  PostgreSQL (pgx)                 | px
//	  Oracle (godror)                  | gr
//	  ---------------------------------|--------------------------------------------
//		 Alibaba MaxCompute (maxcompute)  | mc
//		 Alibaba Tablestore (ots)         | ot, ots, tablestore
//		 Apache Avatica (avatica)         | av, phoenix
//		 Apache H2 (h2)                   | h2
//		 Apache Hive (hive)               | hi
//		 Apache Ignite (ignite)           | ig, gridgain
//		 Apache Impala (impala)           | im
//		 AWS Athena (awsathena)           | s3, aws, athena
//		 Azure Cosmos (cosmos)            | cm
//		 Cassandra (cql)                  | ca, cassandra, datastax, scy, scylla
//		 ClickHouse (clickhouse)          | ch
//		 Couchbase (n1ql)                 | n1, couchbase
//		 Cznic QL (ql)                    | ql, cznic, cznicql
//		 CSVQ (csvq)                      | csv, tsv, json
//		 Exasol (exasol)                  | ex, exa
//		 Databend (databend)              | bend
//		 Firebird SQL (firebirdsql)       | fb, firebird
//		 Genji (genji)                    | gj
//		 Google BigQuery (bigquery)       | bq
//		 Google Spanner (spanner)         | sp
//		 IBM Netezza (nzgo)               | nz, nzgo
//		 Microsoft ADODB (adodb)          | ad, ado
//		 ModernC SQLite (moderncsqlite)   | mq, modernsqlite
//		 ODBC (odbc)                      | od
//		 OLE ODBC (oleodbc)               | oo, ole, oleodbc [adodb]
//		 Presto (presto)                  | pr, prestodb, prestos, prs, prestodbs
//		 SAP ASE (tds)                    | ax, ase, sapase
//		 SAP HANA (hdb)                   | sa, saphana, sap, hana
//		 Snowflake (snowflake)            | sf
//		 Trino (trino)                    | tr, trino, trinos, trs
//		 Vertica (vertica)                | ve
//		 VoltDB (voltdb)                  | vo, volt, vdb
//
// Any protocol scheme alias:// can be used in place of protocol://, and will
// work identically with Parse and Open.
//
// # Using
//
// Please note that the dburl package does not import actual SQL drivers, and
// only provides a standard way to parse/open respective database connection URLs.
//
// For reference, these are the following "expected" SQL drivers that would need
// to be imported:
//
//	  Database (scheme/driver)         | Package
//		----------------------------------|-------------------------------------------------
//		 Microsoft SQL Server (sqlserver) | github.com/microsoft/go-mssqldb
//		 MySQL (mysql)                    | github.com/go-sql-driver/mysql
//		 Oracle Database (oracle)         | github.com/sijms/go-ora/v2
//		 PostgreSQL (postgres)            | github.com/lib/pq
//		 SQLite3 (sqlite3)                | github.com/mattn/go-sqlite3
//	  ---------------------------------|-------------------------------------------------
//		 Amazon Redshift (redshift)       | github.com/lib/pq
//		 CockroachDB (cockroachdb)        | github.com/lib/pq
//		 MemSQL (memsql)                  | github.com/go-sql-driver/mysql
//		 TiDB (tidb)                      | github.com/go-sql-driver/mysql
//		 Vitess (vitess)                  | github.com/go-sql-driver/mysql
//	  ---------------------------------|-------------------------------------------------
//		 MySQL (mymysql)                  | github.com/ziutek/mymysql/godrv
//		 Oracle Database (godror)         | github.com/godror/godror
//		 PostgreSQL (pgx)                 | github.com/jackc/pgx/stdlib
//	  ---------------------------------|-------------------------------------------------
//		 Alibaba MaxCompute (maxcompute)  | sqlflow.org/gomaxcompute
//		 Alibaba Tablestore (ots)  	  | github.com/aliyun/aliyun-tablestore-go-sql-driver
//		 Apache Avatica (avatica)         | github.com/Boostport/avatica
//		 Apache H2 (h2)                   | github.com/jmrobles/h2go
//		 Apache Hive (hive)               | sqlflow.org/gohive
//		 Apache Ignite (ignite)           | github.com/amsokol/ignite-go-client/sql
//		 Apache Impala (impala)           | github.com/bippio/go-impala
//		 AWS Athena (awsathena)           | github.com/uber/athenadriver/go
//		 Azure Cosmos (cosmos)            | github.com/btnguyen2k/gocosmos
//		 Cassandra (cql)                  | github.com/MichaelS11/go-cql-driver
//		 ClickHouse (clickhouse)          | github.com/ClickHouse/clickhouse-go
//		 Couchbase (n1ql)                 | github.com/couchbase/go_n1ql
//		 Cznic QL (ql)                    | modernc.org/ql
//		 CSVQ (csvq)                      | github.com/mithrandie/csvq
//		 Databend (databend)              | github.com/datafuselabs/databend
//		 Exasol (exasol)                  | github.com/exasol/exasol-driver-go
//		 Firebird SQL (firebirdsql)       | github.com/nakagami/firebirdsql
//		 Genji (genji)                    | github.com/genjidb/genji/sql/driver
//		 Google BigQuery (bigquery)       | gorm.io/driver/bigquery/driver
//		 Google Spanner (spanner)         | github.com/rakyll/go-sql-driver-spanner
//		 IBM Netezza (nzgo)               | github.com/IBM/nzgo
//		 Microsoft ADODB (adodb)          | github.com/mattn/go-adodb
//		 ModernC SQLite (moderncsqlite)   | modernc.org/sqlite
//		 ODBC (odbc)                      | github.com/alexbrainman/odbc
//		 OLE ODBC (oleodbc)*              | github.com/mattn/go-adodb
//		 Presto (presto)                  | github.com/prestodb/presto-go-client/presto
//		 SAP ASE (tds)                    | github.com/thda/tds
//		 SAP HANA (hdb)                   | github.com/SAP/go-hdb/driver
//		 Snowflake (snowflake)            | github.com/snowflakedb/gosnowflake
//		 Trino (trino)                    | github.com/trinodb/trino-go-client/trino
//		 Vertica (vertica)                | github.com/vertica/vertica-sql-go
//		 VoltDB (voltdb)                  | github.com/VoltDB/voltdb-client-go/voltdbclient
//
// * OLE ODBC is a special alias for using the "MSDASQL.1" OLE provider with the
// ADODB driver on Windows. oleodbc:// URLs will be converted to the equivalent
// ADODB DSN with "Extended Properties" having the respective ODBC parameters,
// including the underlying transport protocol. As such, oleodbc+transport://user:pass@host/dbname
// URLs are equivalent to adodb://MSDASQL.1/?Extended+Properties=.... on
// Windows. See GenOLEODBC for information regarding how URL components are
// mapped and passed to ADODB's Extended Properties parameter.
//
// # URL Parsing Rules
//
// Parse and Open rely heavily on the standard net/url.URL type, as such
// parsing rules have the same conventions/semantics as any URL parsed by the
// standard library's net/url.Parse.
//
// # Related Projects
//
// This package was written mainly to support these projects:
//
//	https://github.com/xo/usql - a universal command-line interface for SQL databases
//	https://github.com/xo/xo - a command-line tool to generate code for SQL databases
package dburl

import (
	"database/sql"
	"net/url"
	"strings"
)

// Open takes a URL like "protocol+transport://user:pass@host/dbname?option1=a&option2=b"
// and opens a standard sql.DB connection.
//
// See Parse for information on formatting URLs to work properly with Open.
func Open(urlstr string) (*sql.DB, error) {
	u, err := Parse(urlstr)
	if err != nil {
		return nil, err
	}
	return sql.Open(u.Driver, u.DSN)
}

// URL wraps the standard net/url.URL type, adding OriginalScheme, Transport,
// Driver, Unaliased, and DSN strings.
type URL struct {
	// URL is the base net/url/URL.
	url.URL
	// OriginalScheme is the original parsed scheme (ie, "sq", "mysql+unix", "sap", etc).
	OriginalScheme string
	// Transport is the specified transport protocol (ie, "tcp", "udp",
	// "unix", ...), if provided.
	Transport string
	// Driver is the non-aliased SQL driver name that should be used in a call
	// to sql/Open.
	Driver string
	// Unaliased is the unaliased driver name.
	Unaliased string
	// DSN is the built connection "data source name" that can be used in a
	// call to sql/Open.
	DSN string
	// hostPortDB will be set by Gen*() funcs after determining the host, port,
	// database.
	//
	// when empty, indicates that these values are not special, and can be
	// retrieved as the host, port, and path[1:] as usual.
	hostPortDB []string
}

// Parse parses a URL, similar to the standard url.Parse.
//
// Handles parsing OriginalScheme, Transport, Driver, Unaliased, and DSN
// fields.
//
// Note: if the URL has a Opaque component (ie, URLs not specified as
// "scheme://" but "scheme:"), and the database scheme does not support opaque
// components, Parse will attempt to re-process the URL as "scheme://<opaque>".
func Parse(urlstr string) (*URL, error) {
	// parse url
	v, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	if v.Scheme == "" {
		return nil, ErrInvalidDatabaseScheme
	}
	// create url
	u := &URL{
		URL:            *v,
		OriginalScheme: urlstr[:len(v.Scheme)],
		Transport:      "tcp",
	}
	// check for +transport in scheme
	var checkTransport bool
	if i := strings.IndexRune(u.Scheme, '+'); i != -1 {
		u.Transport = urlstr[i+1 : len(v.Scheme)]
		u.Scheme = u.Scheme[:i]
		checkTransport = true
	}
	// get dsn generator
	scheme, ok := schemeMap[u.Scheme]
	if !ok {
		return nil, ErrUnknownDatabaseScheme
	}
	// if scheme does not understand opaque URLs, retry parsing after building
	// fully qualified URL
	if !scheme.Opaque && u.Opaque != "" {
		var q string
		if u.RawQuery != "" {
			q = "?" + u.RawQuery
		}
		var f string
		if u.Fragment != "" {
			f = "#" + u.Fragment
		}
		return Parse(u.OriginalScheme + "://" + u.Opaque + q + f)
	}
	if scheme.Opaque && u.Opaque == "" {
		// force Opaque
		u.Opaque, u.Host, u.Path, u.RawPath = u.Host+u.Path, "", "", ""
	} else if u.Host == "." || (u.Host == "" && strings.TrimPrefix(u.Path, "/") != "") {
		// force unix proto
		u.Transport = "unix"
	}
	// check proto
	if checkTransport || u.Transport != "tcp" {
		if scheme.Transport == TransportNone {
			return nil, ErrInvalidTransportProtocol
		}
		switch {
		case scheme.Transport&TransportAny != 0 && u.Transport != "",
			scheme.Transport&TransportTCP != 0 && u.Transport == "tcp",
			scheme.Transport&TransportUDP != 0 && u.Transport == "udp",
			scheme.Transport&TransportUnix != 0 && u.Transport == "unix":
		default:
			return nil, ErrInvalidTransportProtocol
		}
	}
	// set driver
	u.Driver, u.Unaliased = scheme.Driver, scheme.Driver
	if scheme.Override != "" {
		u.Driver = scheme.Override
	}
	// generate dsn
	if u.DSN, err = scheme.Generator(u); err != nil {
		return nil, err
	}
	return u, nil
}

// String satisfies the stringer interface.
func (u *URL) String() string {
	p := &url.URL{
		Scheme:   u.OriginalScheme,
		Opaque:   u.Opaque,
		User:     u.User,
		Host:     u.Host,
		Path:     u.Path,
		RawPath:  u.RawPath,
		RawQuery: u.RawQuery,
		Fragment: u.Fragment,
	}
	return p.String()
}

// Short provides a short description of the user, host, and database.
func (u *URL) Short() string {
	if u.Scheme == "" {
		return ""
	}
	s := schemeMap[u.Scheme].Aliases[0]
	if u.Scheme == "odbc" || u.Scheme == "oleodbc" {
		n := u.Transport
		if v, ok := schemeMap[n]; ok {
			n = v.Aliases[0]
		}
		s += "+" + n
	} else if u.Transport != "tcp" {
		s += "+" + u.Transport
	}
	s += ":"
	if u.User != nil {
		if n := u.User.Username(); n != "" {
			s += n + "@"
		}
	}
	if u.Host != "" {
		s += u.Host
	}
	if u.Path != "" && u.Path != "/" {
		s += u.Path
	}
	if u.Opaque != "" {
		s += u.Opaque
	}
	return s
}

// Normalize returns the driver, host, port, database, and user name of a URL,
// joined with sep, populating blank fields with empty.
func (u *URL) Normalize(sep, empty string, cut int) string {
	s := []string{u.Unaliased, "", "", "", ""}
	if u.Transport != "tcp" && u.Transport != "unix" {
		s[0] += "+" + u.Transport
	}
	// set host port dbname fields
	if u.hostPortDB == nil {
		if u.Opaque != "" {
			u.hostPortDB = []string{u.Opaque, "", ""}
		} else {
			u.hostPortDB = []string{u.Hostname(), u.Port(), strings.TrimPrefix(u.Path, "/")}
		}
	}
	copy(s[1:], u.hostPortDB)
	// set user
	if u.User != nil {
		s[4] = u.User.Username()
	}
	// replace blank entries ...
	for i := 0; i < len(s); i++ {
		if s[i] == "" {
			s[i] = empty
		}
	}
	if cut > 0 {
		// cut to only populated fields
		i := len(s) - 1
		for ; i > cut; i-- {
			if s[i] != "" {
				break
			}
		}
		s = s[:i]
	}
	return strings.Join(s, sep)
}

// Error is an error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrInvalidDatabaseScheme is the invalid database scheme error.
	ErrInvalidDatabaseScheme Error = "invalid database scheme"
	// ErrUnknownDatabaseScheme is the unknown database type error.
	ErrUnknownDatabaseScheme Error = "unknown database scheme"
	// ErrInvalidTransportProtocol is the invalid transport protocol error.
	ErrInvalidTransportProtocol Error = "invalid transport protocol"
	// ErrRelativePathNotSupported is the relative paths not supported error.
	ErrRelativePathNotSupported Error = "relative path not supported"
	// ErrMissingHost is the missing host error.
	ErrMissingHost Error = "missing host"
	// ErrMissingPath is the missing path error.
	ErrMissingPath Error = "missing path"
	// ErrMissingUser is the missing user error.
	ErrMissingUser Error = "missing user"
)
