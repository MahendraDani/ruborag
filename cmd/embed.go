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

var embedCmd = &cobra.Command{
	Use:   "embed [-w] <input_path>...",
	Short: "Generate vector embeddings of one or more files",
	Long: `The embed command reads parsed text files (produced by ruborag parse)
and computes vector embeddings for each file. When the -w flag is provided,
the embeddings are stored in a local SQLite database (ruborag.db) for later retrieval.

Examples:

  # Embed a single parsed file
  ruborag embed parsed/book.txt

  # Embed a single parsed file and write to index
  ruborag embed -w parsed/book.txt

  # Embed multiple parsed files
  ruborag embed parsed/chapter1.txt parsed/chapter2.txt

  # Embed multiple parsed files and write to index
  ruborag embed -w parsed/chapter1.txt parsed/chapter2.txt

  # Embed all files in a directory
  ruborag embed parsed/

  # Embed all files in a directory and write to index
  ruborag embed -w parsed/
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

func init() {
	rootCmd.AddCommand(embedCmd)
	embedCmd.Flags().BoolVarP(&writeToIndex, "write", "w", false, "Write embeddings to index (SQLite)")
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

	const chunkSize = 1000 //
	chunks := splitTextIntoChunks(text, chunkSize)

	for i, chunk := range chunks {
		vec, err := embedding.EmbedChunk(chunk)
		if err != nil {
			return fmt.Errorf("embedding error for chunk %d of %s: %w", i, path, err)
		}

		if database != nil {
			if err := database.InsertEmbedding(
				filepath.Base(path),
				i, // chunk_index
				chunk,
				vec,
			); err != nil {
				return fmt.Errorf("failed to insert embedding for chunk %d of %s: %w", i, path, err)
			}
			fmt.Printf("stored embedding for %s (chunk %d/%d)\n", path, i+1, len(chunks))
		} else {
			fmt.Printf("embedded %s (chunk %d/%d, %d dimensions)\n", path, i+1, len(chunks), len(vec))
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
