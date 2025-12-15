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

// EmbedClient defines only the methods we need from Gemini client
type EmbedClient interface {
	EmbedContent(
		ctx context.Context,
		model string,
		contents []*genai.Content,
		options *genai.EmbedContentConfig,
	) (*genai.EmbedContentResponse, error)
}

type GeminiClient struct {
	client *genai.Client
}

func (g *GeminiClient) EmbedContent(ctx context.Context, model string, contents []*genai.Content, options *genai.EmbedContentConfig) (*genai.EmbedContentResponse, error) {
	return g.client.Models.EmbedContent(ctx, model, contents, options)
}

func EmbedWithClient(inputPath string, client EmbedClient) []byte {
	text, err := readTextFileBuffered(inputPath)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result, err := client.EmbedContent(
		context.Background(),
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

	return embeddings
}

// returns embeddings from provided filepath
func Embed(inputPath string) []byte {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatalf("Use your own GEMINI_API_KEY, as export GEMINI_API_KEY=<your_key>")
	}
	return EmbedWithClient(inputPath, &GeminiClient{client: client})
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
