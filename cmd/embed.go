package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"ruborag/internal/db"
	"ruborag/internal/embedding"
	"strings"

	"github.com/spf13/cobra"
)

var writeToIndex bool
var useChunking bool
var chunkSize int

var embedCmd = &cobra.Command{
	Use:   "embed [-w|-c] <input_path>...",
	Short: "Generate vector embeddings of one or more files",
	Long: `The embed command reads parsed text files (produced by ruborag parse)
and computes vector embeddings for each file. When the -w flag is provided,
the embeddings are stored in a local SQLite database (ruborag.db) for later retrieval.

By default, each file is embedded as a single unit. When chunking is enabled
using the -c flag, files are split into smaller chunks and each chunk is
embedded independently.

When the -w flag is provided, embeddings are stored in a local SQLite database
(ruborag.db) for later retrieval by the search and ask commands. Otherwise,
embedding information is printed to stdout.

Chunking is recommended for large files, as it improves retrieval quality
and avoids model input size limits.

Options:
  -w, --write            Store embeddings in SQLite index
  -c, --chunk            Enable chunking before embedding
      --chunk-size int   Size of each chunk in characters (default: 1000)

Examples:

  # Embed a single file without chunking
  ruborag embed parsed/book.txt

  # Embed and store embeddings in SQLite
  ruborag embed -w parsed/book.txt

  # Embed with chunking (default chunk size)
  ruborag embed -c parsed/book.txt

  # Embed with chunking and custom chunk size
  ruborag embed -c --chunk-size 2000 parsed/book.txt

  # Embed all parsed files in a directory with chunking and storage
  ruborag embed -w -c parsed/
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("no input files or directories provided")
		}

		var database *db.DB
		var err error

		if writeToIndex {
			database, err = db.Open(db.DefaultDBName)
			if err != nil {
				log.Fatalf("failed to open database: %v", err)
			}
			defer database.Close()
		}

		for _, inputPath := range args {
			if err := processEmbedPath(inputPath, database); err != nil {
				log.Fatalf("embedding failed for %s: %v", inputPath, err)
			}
		}
	},
}

// handles a single file or directory
func processEmbedPath(path string, database *db.DB) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(info.Name(), ".txt") {
				return embedFile(p, database)
			}
			return nil
		})
	}

	return embedFile(path, database)
}

// generates an embedding for a file and writes it to DB if requested
func embedFile(path string, database *db.DB) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	text := string(data)

	var chunks []string
	if useChunking {
		chunks = splitTextIntoChunks(text, chunkSize)
	} else {
		chunks = []string{text} // single chunk = whole file
	}

	sourceFile := filepath.Base(path)

	for i, chunk := range chunks {

		if database != nil {
			exists, err := database.EmbeddingExists(sourceFile, i)
			if err != nil {
				return fmt.Errorf("failed to check existing embedding for %s (chunk %d): %w", path, i, err)
			}

			// Non-chunked mode → skip entire file immediately
			if exists && !useChunking {
				fmt.Printf("skipping %s (already embedded)\n", path)
				return nil
			}

			// Chunked mode → skip only this chunk
			if exists {
				fmt.Printf("skipping %s (chunk %d already embedded)\n", path, i)
				continue
			}
		}

		vec, err := embedding.EmbedChunk(chunk)
		if err != nil {
			return fmt.Errorf("embedding error for chunk %d of %s: %w", i, path, err)
		}

		if database != nil {
			if err := database.InsertEmbedding(
				sourceFile,
				i, // chunk_index
				chunk,
				vec,
			); err != nil {
				return fmt.Errorf("failed to insert embedding for chunk %d of %s: %w", i, path, err)
			}

			fmt.Printf(
				"stored embedding for %s (chunk %d/%d)\n",
				path,
				i+1,
				len(chunks),
			)
		} else {
			fmt.Printf(
				"embedded %s (chunk %d/%d, %d dimensions)\n",
				path,
				i+1,
				len(chunks),
				len(vec),
			)
		}
	}

	return nil
}

func splitTextIntoChunks(text string, chunkSize int) []string {
	var chunks []string
	runes := []rune(text) // handle UTF-8 safely

	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}

	return chunks
}

func init() {
	rootCmd.AddCommand(embedCmd)
	embedCmd.Flags().BoolVarP(&writeToIndex, "write", "w", false, "Write embeddings to index (SQLite)")
	embedCmd.Flags().BoolVarP(&useChunking, "chunk", "c", false, "Enable chunking of files for embeddings")
	embedCmd.Flags().IntVar(&chunkSize, "chunk-size", 1000, "Size of each chunk (in characters) when chunking is enabled")
}
