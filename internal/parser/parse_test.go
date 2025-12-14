package parse_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	parse "ruborag/internal/parser"
)

const FileWritePerm = 0666

const InputHTML = `<p>Welcome to <em>The Rust Programming Language</em>, an introductory book about Rust.
The Rust programming language helps you write faster, more reliable software.
High-level ergonomics and low-level control are often at odds in programming
language design; Rust challenges that conflict. Through balancing powerful
technical capacity and a great developer experience, Rust gives you the option
to control low-level details (such as memory usage) without all the hassle
traditionally associated with such control.</p>`

func TestRemoveHTMLTagsFromFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "temp.html")

	if err := os.WriteFile(filePath, []byte(InputHTML), FileWritePerm); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	output, err := parse.RemoveHTMLTagsFromFile(filePath)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	// Should not contain HTML tags
	if strings.ContainsAny(output, "<>") {
		t.Fatalf("output contains HTML tags: %q", output)
	}

	// Should not contain newlines or tabs
	if strings.ContainsAny(output, "\n\t\r") {
		t.Fatalf("output contains newline or control characters: %q", output)
	}

	// Words should be space-separated (no double spaces)
	if strings.Contains(output, "  ") {
		t.Fatalf("output contains multiple consecutive spaces: %q", output)
	}

	// Sanity check: core content exists
	expectedFragments := []string{
		"Welcome to The Rust Programming Language",
		"Rust gives you the option to control low-level details",
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected fragment missing: %q", fragment)
		}
	}
}
