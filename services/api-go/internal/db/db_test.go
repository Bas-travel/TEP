package db

import (
	"reflect"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	got := splitStatements("-- c\nCREATE TABLE a (\n id INT\n);\n\nINSERT INTO a VALUES (1);\n")
	want := []string{"CREATE TABLE a (\n id INT\n);", "INSERT INTO a VALUES (1);"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
}
