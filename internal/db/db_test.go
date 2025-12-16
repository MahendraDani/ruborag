package db_test

import (
	"os"
	"path/filepath"
	"ruborag/internal/db"
	"testing"
)

func TestOpenCreatesDatabaseAndSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, db.DefaultDBName)

	db, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected db file to exist: %v", err)
	}

	// Verify embeddings table exists
	row := db.QueryRow(`
	SELECT name
	FROM sqlite_master
	WHERE type='table' AND name='embeddings';
`)

	var tableName string
	if err := row.Scan(&tableName); err != nil {
		t.Fatalf("embeddings table does not exist: %v", err)
	}

	if tableName != "embeddings" {
		t.Fatalf("expected table 'embeddings', got %q", tableName)
	}
}
