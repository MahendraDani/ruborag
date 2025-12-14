package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ruborag",
	Short: "A minimal RAG tool for querying The Rust Programming Language book",
	Long: `ruborag is a command-line tool that implements a minimal, explicit
Retrieval-Augmented Generation (RAG) pipeline over The Rust Programming Language book.

The project is designed to demonstrate how RAG works from first principles:
documents are parsed, chunked, embedded, retrieved using cosine similarity,
and injected into an LLM prompt to generate grounded answers.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
