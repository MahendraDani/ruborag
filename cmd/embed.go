package cmd

import (
	"fmt"
	"ruborag/internal/embedding"

	"github.com/spf13/cobra"
)

var writeToIndex bool

var embedCmd = &cobra.Command{
	Use:   "embed [--write] <input_path>...",
	Short: "Generate vector embeddings of one or more files",
	Long: `The embed command reads parsed text files (produced by ruborag parse)
and computes vector embeddings for each text chunk. The embeddings are
stored in a local SQLite database (ruborag.db) for later retrieval by the
search and ask commands.

By default, the command processes files or directories containing parsed
text. Each file is chunked, embedded, and stored in the database with its
source and chunk index.

Examples:

  # Embed a single parsed file
  ruborag embed parsed/book.txt

	# Embed a single parsed file and write to index
  ruborag embed -w parsed/book.txt

  # Embed multiple parsed files
  ruborag embed parsed/chapter1.txt parsed/chapter2.txt

	# Embed multiple parsed files
  ruborag embed -w parsed/chapter1.txt parsed/chapter2.txt

  # Embed all files in a directory
  ruborag embed parsed/

	# Embed all files in a directory and write to index
  ruborag embed -w parsed/
	
`,
	Run: func(cmd *cobra.Command, args []string) {
		embeddings := embedding.Embed(args[0])
		fmt.Println(string(embeddings))
	},
}

func init() {
	rootCmd.AddCommand(embedCmd)

	parseCmd.Flags().BoolVarP(&writeToFile, "write", "w", false, "Write output to index")
}
