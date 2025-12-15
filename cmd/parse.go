package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"ruborag/internal/parser"
	"strings"

	"github.com/spf13/cobra"
)

var writeToFile bool

var parseCmd = &cobra.Command{
	Use:   "parse [-w] <input_file>...",
	Short: "Parse one or more input files",
	Long: `Parse one or more input files and extract their readable content.

By default, the parsed output is written to stdout. When the -w flag is
provided, the output of each input file is written to a corresponding
output file with "-parsed" appended to the filename.

The command accepts one or more input files.

Examples:
  ruborag parse book.html
  ruborag parse chapter1.html chapter2.html
  ruborag parse -w book.html
  ruborag parse -w chapter1.html chapter2.html`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			cmd.Usage()
			return
		}
		// check if given args is a html file, parse it
		// if given file is a directory, parse every html file in the directory
		for _, path := range args {
			info, err := os.Stat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error accessing %s: %v\n", path, err)
				continue
			}

			if info.IsDir() {
				err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}

					if d.IsDir() {
						return nil
					}

					if filepath.Ext(p) != ".html" {
						return nil
					}

					return parseSingleFile(p)
				})

				if err != nil {
					fmt.Fprintf(os.Stderr, "error parsing directory %s: %v\n", path, err)
				}
			} else {
				if filepath.Ext(path) != ".html" {
					fmt.Fprintf(os.Stderr, "skipping non-html file: %s\n", path)
					continue
				}
				if err := parseSingleFile(path); err != nil {
					fmt.Fprintf(os.Stderr, "error parsing file %s: %v\n", path, err)
				}
			}
		}
	},
}

func parseSingleFile(inputPath string) error {
	content, err := parser.RemoveHTMLTagsFromFile(inputPath)
	if err != nil {
		return err
	}

	if writeToFile {
		outPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + "-parsed.txt"
		return os.WriteFile(outPath, []byte(content), 0644)
	}

	fmt.Println(content)
	return nil
}

func init() {
	rootCmd.AddCommand(parseCmd)
	parseCmd.Flags().BoolVarP(&writeToFile, "write", "w", false, "Write output to a file")
}
