package main

import (
	"strings"
	"testing"
)

func TestSQLiteDSNAddsConnectionPragmas(t *testing.T) {
	dsn := sqliteDSN("/data/db/dogspots.db")
	for _, pragma := range []string{"_pragma=busy_timeout(5000)", "_pragma=foreign_keys(1)", "_pragma=journal_mode(WAL)"} {
		if !strings.Contains(dsn, pragma) {
			t.Errorf("sqliteDSN() = %q, want it to contain %q", dsn, pragma)
		}
	}
}

func TestSQLiteDSNAppendsWithAmpersand(t *testing.T) {
	dsn := sqliteDSN("file:dogspots.db?cache=shared")
	if !strings.Contains(dsn, "?cache=shared&_pragma=") {
		t.Fatalf("sqliteDSN() = %q, want the pragmas appended with &", dsn)
	}
}
