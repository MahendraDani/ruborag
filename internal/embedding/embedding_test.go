package embedding

import (
	"os"
	"path/filepath"
	"testing"
)

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
