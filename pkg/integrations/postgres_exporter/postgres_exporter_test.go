package postgres_exporter //nolint:golint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parsePostgresURL(t *testing.T) {
	url := "postgresql://user:'pass'@localhost:5432/postgres?sslmode=disable"

	kvp, err := parsePostgresURL(url)
	require.NoError(t, err)

	require.Equal(t, "postgres", kvp["dbname"])
	require.Equal(t, "localhost", kvp["host"])
	require.Equal(t, "'pass'", kvp["password"])
	require.Equal(t, "5432", kvp["port"])
	require.Equal(t, "disable", kvp["sslmode"])
	require.Equal(t, "user", kvp["user"])
}
