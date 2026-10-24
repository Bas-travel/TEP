// Package db opens MySQL, runs embedded migrations, and pings Redis (RESP) without extra dependencies.
package db

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	d, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(20)
	d.SetMaxIdleConns(5)
	d.SetConnMaxLifetime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d, d.PingContext(ctx)
}

// splitStatements splits a SQL script on semicolons that end a line (no ';' inside string literals in our files).
func splitStatements(script string) []string {
	var out []string
	var cur strings.Builder
	for _, line := range strings.Split(script, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "--") {
			continue
		}
		cur.WriteString(line + "\n")
		if strings.HasSuffix(t, ";") {
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
		}
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		out = append(out, s)
	}
	return out
}

// Migrate applies pending *.up.sql files in order. Files whose name contains "_seed_" only run when seed is true.
// Returns the versions applied by this call.
func Migrate(ctx context.Context, d *sql.DB, files fs.FS, seed bool) ([]string, error) {
	if _, err := d.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(64) NOT NULL PRIMARY KEY, applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return nil, err
	}
	names, err := fs.Glob(files, "*.up.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	var applied []string
	for _, n := range names {
		ver := strings.TrimSuffix(n, ".up.sql")
		if strings.Contains(ver, "_seed_") && !seed {
			continue
		}
		var x string
		err := d.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE version=?`, ver).Scan(&x)
		if err == nil {
			continue
		}
		if err != sql.ErrNoRows {
			return applied, err
		}
		raw, err := fs.ReadFile(files, n)
		if err != nil {
			return applied, err
		}
		for i, st := range splitStatements(string(raw)) {
			if _, err := d.ExecContext(ctx, st); err != nil {
				return applied, fmt.Errorf("migration %s statement %d: %w", ver, i+1, err)
			}
		}
		if _, err := d.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES(?)`, ver); err != nil {
			return applied, err
		}
		applied = append(applied, ver)
	}
	return applied, nil
}

// PingRedis sends an inline PING and expects +PONG.
func PingRedis(ctx context.Context, addr string) error {
	var dl net.Dialer
	c, err := dl.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := c.Write([]byte("PING\r\n")); err != nil {
		return err
	}
	line, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, "+PONG") {
		return fmt.Errorf("unexpected redis reply %q", strings.TrimSpace(line))
	}
	return nil
}
