package db

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const DefaultDBName = "ruboragdb"

type DB struct {
	conn *sql.DB
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping sqlite db: %w", err)
	}

	db := &DB{conn: conn}

	if _, err := db.conn.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *DB) Exec(query string, args ...any) error {
	_, err := db.conn.Exec(query, args...)
	return err
}

func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

func (db *DB) Close() error {
	if db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

func (db *DB) initSchema() error {
	const schema = `
	CREATE TABLE IF NOT EXISTS embeddings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_file TEXT NOT NULL,
		chunk_index INTEGER NOT NULL,
		content TEXT NOT NULL,
		embedding BLOB NOT NULL
	);
	`
	if _, err := db.conn.Exec(schema); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

func (db *DB) InsertEmbedding(
	sourceFile string,
	chunkIndex int,
	content string,
	embedding []float32,
) error {
	if len(embedding) == 0 {
		return fmt.Errorf("embedding cannot be empty")
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, embedding); err != nil {
		return fmt.Errorf("encode embedding: %w", err)
	}

	const query = `
	INSERT INTO embeddings (
		source_file,
		chunk_index,
		content,
		embedding
	) VALUES (?, ?, ?, ?);
	`

	_, err := db.conn.Exec(
		query,
		sourceFile,
		chunkIndex,
		content,
		buf.Bytes(),
	)
	if err != nil {
		return fmt.Errorf("insert embedding: %w", err)
	}

	return nil
}
