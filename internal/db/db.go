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

func (db *DB) EmbeddingExists(sourceFile string, chunkIndex int) (bool, error) {
	const query = `
	SELECT 1
	FROM embeddings
	WHERE source_file = ? AND chunk_index = ?
	LIMIT 1;
	`

	var dummy int
	err := db.conn.QueryRow(query, sourceFile, chunkIndex).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

type StoredEmbedding struct {
	SourceFile string
	ChunkIndex int
	Vector     []float32
}

func (db *DB) GetAllEmbeddings() ([]StoredEmbedding, error) {
	const query = `
	SELECT source_file, chunk_index, embedding
	FROM embeddings;
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query embeddings: %w", err)
	}
	defer rows.Close()

	var results []StoredEmbedding

	for rows.Next() {
		var sourceFile string
		var chunkIndex int
		var blob []byte

		if err := rows.Scan(&sourceFile, &chunkIndex, &blob); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		vec, err := DecodeEmbedding(blob)
		if err != nil {
			return nil, fmt.Errorf("decode embedding: %w", err)
		}

		results = append(results, StoredEmbedding{
			SourceFile: sourceFile,
			ChunkIndex: chunkIndex,
			Vector:     vec,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func DecodeEmbedding(blob []byte) ([]float32, error) {
	if len(blob)%4 != 0 {
		return nil, fmt.Errorf("invalid embedding blob size")
	}

	vec := make([]float32, len(blob)/4)
	reader := bytes.NewReader(blob)

	if err := binary.Read(reader, binary.LittleEndian, &vec); err != nil {
		return nil, err
	}

	return vec, nil
}
