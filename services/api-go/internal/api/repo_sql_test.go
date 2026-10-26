package api

import (
	"context"
	"testing"

	"github.com/Bas-travel/tep/services/api-go/internal/db"
	"github.com/Bas-travel/tep/services/api-go/migrations"
)

// catalogTables hold seeded reference data and survive a reset; every other table is user-level and is emptied.
var catalogTables = map[string]bool{"destinations": true, "listings": true, "faqs": true, "schema_migrations": true}

// sqlTestRepo opens TEP_TEST_MYSQL_DSN, migrates (with seed), and empties all user-level tables so every test starts from
// the same state and ids are deterministic (demo user = 1, posts 1..5, ...).
func sqlTestRepo(t *testing.T, dsn string) Repo {
	t.Helper()
	d, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	ctx := context.Background()
	if _, err := db.Migrate(ctx, d, migrations.FS, true); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	conn, err := d.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	rows, err := conn.QueryContext(ctx, `SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND table_type='BASE TABLE'`)
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var n string
		_ = rows.Scan(&n)
		if !catalogTables[n] {
			tables = append(tables, n)
		}
	}
	rows.Close()
	if _, err := conn.ExecContext(ctx, `SET FOREIGN_KEY_CHECKS=0`); err != nil {
		t.Fatal(err)
	}
	for _, n := range tables {
		if _, err := conn.ExecContext(ctx, "TRUNCATE TABLE `"+n+"`"); err != nil {
			t.Fatalf("truncate %s: %v", n, err)
		}
	}
	_, _ = conn.ExecContext(ctx, `SET FOREIGN_KEY_CHECKS=1`)
	r := NewSQLRepo(d)
	t.Cleanup(func() { _ = r.Close() })
	return r
}
