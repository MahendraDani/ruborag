package embedding

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/genai"
)

type fakeClient struct{}

func (f *fakeClient) EmbedContent(ctx context.Context, model string, contents []*genai.Content, options *genai.EmbedContentConfig) (*genai.EmbedContentResponse, error) {
	return &genai.EmbedContentResponse{
		Embeddings: []*genai.ContentEmbedding{
			{
				Values: []float32{0.1, 0.2, 0.3},
			},
		},
	}, nil

}

func TestEmbedWithClient(t *testing.T) {
	// write a temp file
	path := t.TempDir() + "/test.txt"
	content := "Hello embeddings!"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// call EmbedWithClient
	embeddings := EmbedWithClient(path, &fakeClient{})

	// check that the output contains valid JSON
	if !json.Valid(embeddings) {
		t.Fatalf("output is not valid JSON: %s", string(embeddings))
	}

	// optional: check that it contains the vector we set
	var parsed []*genai.ContentEmbedding
	if err := json.Unmarshal(embeddings, &parsed); err != nil {
		t.Fatalf("failed to unmarshal embeddings: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 embedding, got %d", len(parsed))
	}

	expected := []float32{0.1, 0.2, 0.3}
	for i, v := range parsed[0].Values {
		if v != expected[i] {
			t.Fatalf("expected vector[%d]=%v, got %v", i, expected[i], v)
		}
	}
}

func writeTempFile(t *testing.T, content []byte) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return path
}

func TestReadTextFileBuffered_ValidText(t *testing.T) {
	data := []byte("Hello world\nThis is a test file.")
	path := writeTempFile(t, data)

	text, err := readTextFileBuffered(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if text != string(data) {
		t.Fatalf("expected %q, got %q", string(data), text)
	}
}

func TestReadTextFileBuffered_EmptyFile(t *testing.T) {
	path := writeTempFile(t, []byte{})

	_, err := readTextFileBuffered(path)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

func TestReadTextFileBuffered_BinaryFile(t *testing.T) {
	// Invalid UTF-8 sequence
	binary := []byte{0xff, 0xfe, 0xfd, 0x00}
	path := writeTempFile(t, binary)

	_, err := readTextFileBuffered(path)
	if err == nil {
		t.Fatal("expected error for binary file, got nil")
	}
}

func TestReadTextFileBuffered_FileDoesNotExist(t *testing.T) {
	_, err := readTextFileBuffered("does-not-exist.txt")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestReadTextFileBuffered_LargeFile(t *testing.T) {
	large := make([]byte, 0, 1024*1024)
	for i := 0; i < 100_000; i++ {
		large = append(large, []byte("some text line\n")...)
	}

	path := writeTempFile(t, large)

	text, err := readTextFileBuffered(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(text) != len(large) {
		t.Fatalf("expected length %d, got %d", len(large), len(text))
	}
}
