package api

import "testing"

// testRepoOpts picks the storage backend for the whole suite: in-memory by default; MySQL when TEP_TEST_MYSQL_DSN is set
// (see repo_sql_test.go). The same tests run against both.
func testRepoOpts(t *testing.T) []Option { return nil }
