package postgres_exporter

import (
	"reflect"
	"testing"
)

func Test_ParsePostgresURL(t *testing.T) {
	dsn := "postgresql://linus:42secret@localhost:5432/postgres?sslmode=disable"
	expected := map[string]string{
		"dbname":   "'postgres'",
		"host":     "'localhost'",
		"password": "'42secret'",
		"port":     "'5432'",
		"sslmode":  "'disable'",
		"user":     "'linus'",
	}

	actual, _ := parsePostgresURL(dsn)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("parsePortgresURL failed, actual: %v, expected: %v", actual, expected)
	}

}
