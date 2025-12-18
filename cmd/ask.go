package cmd

import (
	"fmt"
	"log"
	"ruborag/internal/db"
	"ruborag/internal/embedding"
	"ruborag/internal/search"
	"strings"

	"context"

	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask questions grounded in the Rust Book",
	Long: `The ask command answers questions using Retrieval-Augmented Generation (RAG).

It retrieves the most relevant Rust Book passages using semantic search
and provides them as context to an LLM to generate a grounded answer.

Examples:

  ruborag ask "what is borrowing in rust"
  ruborag ask "how do lifetimes work?"

`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := args[0]

		log.Println("Embedding query...")
		queryVec, err := embedding.EmbedChunk(question)
		if err != nil {
			log.Fatalf("failed to embed query: %v", err)
		}

		database, err := db.Open(db.DefaultDBName)
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}
		defer database.Close()

		log.Println("Loading Rust Book embeddings...")
		embeddings, err := database.GetAllEmbeddings()
		if err != nil {
			log.Fatalf("failed to load embeddings: %v", err)
		}

		log.Println("Searching for relevant context...")
		results := search.SearchQuery(queryVec, embeddings)
		if len(results) == 0 {
			log.Fatal("no relevant context found")
		}

		limit := 3
		if len(results) < limit {
			limit = len(results)
		}
		log.Printf("Selected top %d most relevant passages:\n", limit)

		var ctxBuilder strings.Builder
		for i := 0; i < limit; i++ {
			r := results[i]
			log.Printf("  [%d] %s\n", i+1, r.SourceFile)

			fmt.Fprintf(&ctxBuilder,
				"[%d] Source: %s \n%s\n\n",
				i+1,
				r.SourceFile,
				r.Content,
			)
		}

		prompt := fmt.Sprintf(`
You are a Rust programming expert.
Answer the question strictly using the provided context.
If the answer is not present, say you don't know.

CONTEXT:
%s

QUESTION:
%s
`, ctxBuilder.String(), question)

		log.Println("Gemini is thinking...")
		ctx := context.Background()
		client, err := genai.NewClient(ctx, nil)
		if err != nil {
			log.Fatal("GEMINI_API_KEY not set")
		}

		resp, err := client.Models.GenerateContent(
			ctx,
			"gemini-3-flash-preview",
			genai.Text(prompt),
			nil,
		)
		if err != nil {
			log.Fatalf("LLM error: %v", err)
		}

		log.Println("Gemini response generated:\n")
		fmt.Println(resp.Text())
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
