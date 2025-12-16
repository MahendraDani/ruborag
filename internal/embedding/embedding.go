package embedding

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
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

func (g *GeminiClient) EmbedContent(
	ctx context.Context,
	model string,
	contents []*genai.Content,
	options *genai.EmbedContentConfig,
) (*genai.EmbedContentResponse, error) {
	return g.client.Models.EmbedContent(ctx, model, contents, options)
}

// EmbedWithClient generates an embedding vector for the given file
func EmbedWithClient(inputPath string, client EmbedClient) ([]float32, error) {
	text, err := readTextFileBuffered(inputPath)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned from model")
	}

	// Single input → single embedding
	return result.Embeddings[0].Values, nil
}

// Embed creates a Gemini client and generates embeddings for the file
func Embed(inputPath string) ([]float32, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create Gemini client (ensure GEMINI_API_KEY is set): %w",
			err,
		)
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

func EmbedChunk(text string) ([]float32, error) {
	// Wrap text in a temporary in-memory "client" request
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result, err := (&GeminiClient{client: client}).EmbedContent(ctx, "gemini-embedding-001", contents, nil)
	if err != nil {
		return nil, err
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Embeddings[0].Values, nil
}
