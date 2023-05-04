package dburl

import (
	"fmt"
	"sort"
)

// Transport is the allowed transport protocol types in a database URL scheme.
type Transport uint

// Transport types.
const (
	TransportNone Transport = 0
	TransportTCP  Transport = 1
	TransportUDP  Transport = 2
	TransportUnix Transport = 4
	TransportAny  Transport = 8
)

// Scheme wraps information used for registering a database URL scheme for use
// with Parse/Open.
type Scheme struct {
	// Driver is the name of the SQL driver that is set as the Scheme in
	// Parse'd URLs and is the driver name expected by the standard sql.Open
	// calls.
	//
	// Note: a 2 letter alias will always be registered for the Driver as the
	// first 2 characters of the Driver, unless one of the Aliases includes an
	// alias that is 2 characters.
	Driver string
	// Generator is the func responsible for generating a DSN based on parsed
	// URL information.
	//
	// Note: this func should not modify the passed URL.
	Generator func(*URL) (string, error)
	// Transport are allowed protocol transport types for the scheme.
	Transport Transport
	// Opaque toggles Parse to not re-process URLs with an "opaque" component.
	Opaque bool
	// Aliases are any additional aliases for the scheme.
	Aliases []string
	// Override is the Go SQL driver to use instead of Driver.
	//
	// Used for "wire compatible" driver schemes.
	Override string
}

// BaseSchemes returns the supported base schemes.
func BaseSchemes() []Scheme {
	return []Scheme{
		// core databases
		{"mysql", GenMysql, TransportTCP | TransportUDP | TransportUnix, false, []string{"mariadb", "maria", "percona", "aurora"}, ""},
		{"oracle", GenFromURL("oracle://localhost:1521"), 0, false, []string{"ora", "oci", "oci8", "odpi", "odpi-c"}, ""},
		{"postgres", GenPostgres, TransportUnix, false, []string{"pg", "postgresql", "pgsql"}, ""},
		{"sqlite3", GenOpaque, 0, true, []string{"sqlite", "file"}, ""},
		{"sqlserver", GenScheme("sqlserver"), 0, false, []string{"ms", "mssql", "azuresql"}, ""},
		// wire compatibles
		{"cockroachdb", GenFromURL("postgres://localhost:26257/?sslmode=disable"), 0, false, []string{"cr", "cockroach", "crdb", "cdb"}, "postgres"},
		{"memsql", GenMysql, 0, false, nil, "mysql"},
		{"redshift", GenFromURL("postgres://localhost:5439/"), 0, false, []string{"rs"}, "postgres"},
		{"tidb", GenMysql, 0, false, nil, "mysql"},
		{"vitess", GenMysql, 0, false, []string{"vt"}, "mysql"},
		// alternate implementations
		{"godror", GenGodror, 0, false, []string{"gr"}, ""},
		{"moderncsqlite", GenOpaque, 0, true, []string{"mq", "modernsqlite"}, ""},
		{"mymysql", GenMymysql, TransportTCP | TransportUDP | TransportUnix, false, []string{"zm", "mymy"}, ""},
		{"pgx", GenFromURL("postgres://localhost:5432/"), TransportUnix, false, []string{"px"}, ""},
		// other databases
		{"adodb", GenAdodb, 0, false, []string{"ado"}, ""},
		{"awsathena", GenScheme("s3"), 0, false, []string{"s3", "aws", "athena"}, ""},
		{"avatica", GenFromURL("http://localhost:8765/"), 0, false, []string{"phoenix"}, ""},
		{"bigquery", GenScheme("bigquery"), 0, false, []string{"bq"}, ""},
		{"clickhouse", GenFromURL("clickhouse://localhost:9000/"), 0, false, []string{"ch"}, ""},
		{"cosmos", GenCosmos, 0, false, []string{"cm"}, ""},
		{"cql", GenCassandra, 0, false, []string{"ca", "cassandra", "datastax", "scy", "scylla"}, ""},
		{"csvq", GenOpaque, 0, true, []string{"csv", "tsv", "json"}, ""},
		{"databend", GenDatabend, 0, false, []string{"dd", "bend"}, ""},
		{"exasol", GenExasol, 0, false, []string{"ex", "exa"}, ""},
		{"firebirdsql", GenFirebird, 0, false, []string{"fb", "firebird"}, ""},
		{"genji", GenOpaque, 0, true, []string{"gj"}, ""},
		{"h2", GenFromURL("h2://localhost:9092/"), 0, false, nil, ""},
		{"hdb", GenScheme("hdb"), 0, false, []string{"sa", "saphana", "sap", "hana"}, ""},
		{"hive", GenSchemeTruncate, 0, false, nil, ""},
		{"ignite", GenIgnite, 0, false, []string{"ig", "gridgain"}, ""},
		{"impala", GenScheme("impala"), 0, false, nil, ""},
		{"maxcompute", GenSchemeTruncate, 0, false, []string{"mc"}, ""},
		{"n1ql", GenFromURL("http://localhost:9000/"), 0, false, []string{"couchbase"}, ""},
		{"nzgo", GenPostgres, TransportUnix, false, []string{"nz", "netezza"}, ""},
		{"odbc", GenOdbc, TransportAny, false, nil, ""},
		{"oleodbc", GenOleodbc, TransportAny, false, []string{"oo", "ole"}, "adodb"},
		{"ots", GenTableStore, TransportAny, false, []string{"tablestore"}, ""},
		{"presto", GenPresto, 0, false, []string{"prestodb", "prestos", "prs", "prestodbs"}, ""},
		{"ql", GenOpaque, 0, true, []string{"ql", "cznic", "cznicql"}, ""},
		{"snowflake", GenSnowflake, 0, false, []string{"sf"}, ""},
		{"spanner", GenSpanner, 0, false, []string{"sp"}, ""},
		{"tds", GenFromURL("http://localhost:5000/"), 0, false, []string{"ax", "ase", "sapase"}, ""},
		{"trino", GenPresto, 0, false, []string{"trino", "trinos", "trs"}, ""},
		{"vertica", GenFromURL("vertica://localhost:5433/"), 0, false, nil, ""},
		{"voltdb", GenVoltdb, 0, false, []string{"volt", "vdb"}, ""},
	}
}

func init() {
	// register schemes
	schemes := BaseSchemes()
	schemeMap = make(map[string]*Scheme, len(schemes))
	for _, scheme := range schemes {
		Register(scheme)
	}
}

// schemeMap is the map of registered schemes.
var schemeMap map[string]*Scheme

// registerAlias registers a alias for an already registered Scheme.
func registerAlias(name, alias string, doSort bool) {
	scheme, ok := schemeMap[name]
	if !ok {
		panic(fmt.Sprintf("scheme %s not registered", name))
	}
	if doSort && contains(scheme.Aliases, alias) {
		panic(fmt.Sprintf("scheme %s already has alias %s", name, alias))
	}
	if _, ok := schemeMap[alias]; ok {
		panic(fmt.Sprintf("scheme %s already registered", alias))
	}
	scheme.Aliases = append(scheme.Aliases, alias)
	if doSort {
		sort.Slice(scheme.Aliases, func(i, j int) bool {
			if len(scheme.Aliases[i]) <= len(scheme.Aliases[j]) {
				return true
			}
			if len(scheme.Aliases[j]) < len(scheme.Aliases[i]) {
				return false
			}
			return scheme.Aliases[i] < scheme.Aliases[j]
		})
	}
	schemeMap[alias] = scheme
}

// Register registers a Scheme.
func Register(scheme Scheme) {
	if scheme.Generator == nil {
		panic("must specify Generator when registering Scheme")
	}
	if scheme.Opaque && scheme.Transport&TransportUnix != 0 {
		panic("scheme must support only Opaque or Unix protocols, not both")
	}
	// check if registered
	if _, ok := schemeMap[scheme.Driver]; ok {
		panic(fmt.Sprintf("scheme %s already registered", scheme.Driver))
	}
	sz := &Scheme{
		Driver:    scheme.Driver,
		Generator: scheme.Generator,
		Transport: scheme.Transport,
		Opaque:    scheme.Opaque,
		Override:  scheme.Override,
	}
	schemeMap[scheme.Driver] = sz
	// add aliases
	var hasShort bool
	for _, alias := range scheme.Aliases {
		if len(alias) == 2 {
			hasShort = true
		}
		if scheme.Driver != alias {
			registerAlias(scheme.Driver, alias, false)
		}
	}
	if !hasShort && len(scheme.Driver) > 2 {
		registerAlias(scheme.Driver, scheme.Driver[:2], false)
	}
	// ensure always at least one alias, and that if Driver is 2 characters,
	// that it gets added as well
	if len(sz.Aliases) == 0 || len(scheme.Driver) == 2 {
		sz.Aliases = append(sz.Aliases, scheme.Driver)
	}
	// sort
	sort.Slice(sz.Aliases, func(i, j int) bool {
		if len(sz.Aliases[i]) <= len(sz.Aliases[j]) {
			return true
		}
		if len(sz.Aliases[j]) < len(sz.Aliases[i]) {
			return false
		}
		return sz.Aliases[i] < sz.Aliases[j]
	})
}

// Unregister unregisters a Scheme and all associated aliases.
func Unregister(name string) *Scheme {
	if scheme, ok := schemeMap[name]; ok {
		for _, alias := range scheme.Aliases {
			delete(schemeMap, alias)
		}
		delete(schemeMap, name)
		return scheme
	}
	return nil
}

// RegisterAlias registers a alias for an already registered Scheme.
func RegisterAlias(name, alias string) {
	registerAlias(name, alias, true)
}

// Protocols returns list of all valid protocol aliases for a scheme name.
func Protocols(name string) []string {
	if scheme, ok := schemeMap[name]; ok {
		return append([]string{scheme.Driver}, scheme.Aliases...)
	}
	return nil
}

// SchemeDriverAndAliases returns the registered driver and aliases for a
// database scheme.
func SchemeDriverAndAliases(name string) (string, []string) {
	if scheme, ok := schemeMap[name]; ok {
		driver := scheme.Driver
		if scheme.Override != "" {
			driver = scheme.Override
		}
		var aliases []string
		for _, alias := range scheme.Aliases {
			if alias == driver {
				continue
			}
			aliases = append(aliases, alias)
		}
		sort.Slice(aliases, func(i, j int) bool {
			if len(aliases[i]) <= len(aliases[j]) {
				return true
			}
			if len(aliases[j]) < len(aliases[i]) {
				return false
			}
			return aliases[i] < aliases[j]
		})
		return driver, aliases
	}
	return "", nil
}

// ShortAlias returns the short alias for the scheme name.
func ShortAlias(name string) string {
	return schemeMap[name].Aliases[0]
}

// contains determines if v contains s.
func contains(v []string, s string) bool {
	for _, z := range v {
		if z == s {
			return true
		}
	}
	return false
}
