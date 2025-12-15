package embedding

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"unicode/utf8"

	"google.golang.org/genai"
)

// returns embeddings from provided filepath
func Embed(inputPath string) {
	text, err := readTextFileBuffered(inputPath)
	if err != nil {
		log.Printf("failed to open file: %v", err)
	}

	if len(text) == 0 {
		log.Fatalf("file %s is empty", inputPath)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatalf("Use your own GEMINI_API_KEY, as export GEMINI_API_KEY=<your_key>")
	}

	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result, err := client.Models.EmbedContent(
		ctx,
		"gemini-embedding-001",
		contents,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	embeddings, err := json.MarshalIndent(result.Embeddings, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(embeddings))
}

// readTextFileBuffered reads a file using buffered IO and ensures it is UTF-8 text
func readTextFileBuffered(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	reader := bufio.NewReader(file)

	for {
		chunk, err := reader.ReadBytes('\n')
		buf.Write(chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
	}

	content := buf.Bytes()

	if !utf8.Valid(content) {
		return "", fmt.Errorf("file is not valid UTF-8 text")
	}

	if len(content) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	return string(content), nil
}
