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
var outDir string

var parseCmd = &cobra.Command{
	Use:   "parse [--write --out-dir <dir>] <input_path>...",
	Short: "Parse one or more input files",
	Long: `Parse one or more input HTML files and extract their readable text content.

The parse command removes HTML markup and outputs clean, human-readable text.
Input arguments may be individual HTML files or directories containing HTML files.

By default, parsed content is written to stdout, which makes the command
compatible with standard Unix pipelines for exploratory use.

When the --write flag is provided, parsed output is written to files instead
of stdout. In this mode, an output directory must be specified using --out-dir.
Each input HTML file produces a corresponding "-parsed.txt" file in the output
directory.

Examples:
  # Parse a single file and print to stdout
  ruborag parse book.html

  # Parse multiple files
  ruborag parse chapter1.html chapter2.html

  # Parse a directory of HTML files and write output files
  ruborag parse rust-book/ --write --out-dir out-again

  # Use stdout output with Unix tools
  ruborag parse book.html | less
`,
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
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}

		base := filepath.Base(inputPath)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		outPath := filepath.Join(outDir, name+"-parsed.txt")

		return os.WriteFile(outPath, []byte(content), 0644)
	}

	fmt.Println(content)
	return nil
}

func init() {
	rootCmd.AddCommand(parseCmd)
	parseCmd.Flags().BoolVarP(&writeToFile, "write", "w", false, "Write output to a file")
	parseCmd.Flags().StringVar(&outDir, "out-dir", "", "Directory to write parsed output")
}
