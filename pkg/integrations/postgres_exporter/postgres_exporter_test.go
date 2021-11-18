package postgres_exporter //nolint:golint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParsePostgresURL(t *testing.T) {
	dsn := "postgresql://linus:42secret@localhost:5432/postgres?sslmode=disable"
	expected := map[string]string{
		"dbname":   "postgres",
		"host":     "localhost",
		"password": "42secret",
		"port":     "5432",
		"sslmode":  "disable",
		"user":     "linus",
	}

	actual, err := parsePostgresURL(dsn)
	require.NoError(t, err)
	require.Equal(t, actual, expected)

}
