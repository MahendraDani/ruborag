package cmd

import (
	"fmt"
	"log"
	"ruborag/internal/db"
	"ruborag/internal/embedding"
	"ruborag/internal/similarity"
	"sort"

	"github.com/spf13/cobra"
)

var topK int

type searchResult struct {
	SourceFile string
	ChunkIndex int
	Score      float32
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Rust Book content using semantic similarity",
	Long: `The search command performs semantic search over the Rust Book
using vector embeddings stored in the local SQLite index.

It converts the query into an embedding, computes cosine similarity
against all stored embeddings, and returns the most relevant results.

Examples:

  # Basic semantic search
  ruborag search "what is borrowing in rust"

  # Return top 10 results
  ruborag search --top-k 10 "what is ownership"

`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		queryVec, err := embedding.EmbedChunk(query)
		if err != nil {
			log.Fatalf("failed to embed query: %v", err)
		}

		database, err := db.Open(db.DefaultDBName)
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}
		defer database.Close()

		embeddings, err := database.GetAllEmbeddings()
		if err != nil {
			log.Fatalf("failed to load embeddings: %v", err)
		}

		if len(embeddings) == 0 {
			log.Fatal("no embeddings found in database")
		}

		results := make([]searchResult, 0, len(embeddings))

		for _, e := range embeddings {
			score := similarity.CosineSimilarity(queryVec, e.Vector)
			results = append(results, searchResult{
				SourceFile: e.SourceFile,
				ChunkIndex: e.ChunkIndex,
				Score:      score,
			})
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})

		if topK > len(results) {
			topK = len(results)
		}

		fmt.Printf("Top %d results:\n\n", topK)
		for i := 0; i < topK; i++ {
			r := results[i]
			fmt.Printf(
				"%d. %s (chunk %d) — score: %.4f\n",
				i+1,
				r.SourceFile,
				r.ChunkIndex,
				r.Score,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().IntVarP(
		&topK,
		"top-k",
		"k",
		5,
		"Number of top results to return",
	)
}
