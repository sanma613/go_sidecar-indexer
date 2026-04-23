package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestCSVSourceReadAll(t *testing.T) {
	tempDir := t.TempDir()
	path := writeTempFile(t, tempDir, "docs.csv", "id,content\n10,hello csv\n")

	source := NewCSVSource(path)
	documents, err := source.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if len(documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(documents))
	}
	if documents[0].ID != 10 || documents[0].Content != "hello csv" {
		t.Fatalf("unexpected document: %+v", documents[0])
	}
}

func TestJSONSourceReadAll(t *testing.T) {
	tempDir := t.TempDir()
	path := writeTempFile(t, tempDir, "docs.json", `[{"id":2,"content":"hello json"}]`)

	source := NewJSONSource(path)
	documents, err := source.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if len(documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(documents))
	}
	if documents[0].ID != 2 || documents[0].Content != "hello json" {
		t.Fatalf("unexpected document: %+v", documents[0])
	}
}

func TestPDFSourceReadAllMVP(t *testing.T) {
	tempDir := t.TempDir()
	path := writeTempFile(t, tempDir, "docs.pdf", "%PDF-1.4\nHello PDF text")

	source := NewPDFSource(path)
	documents, err := source.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if len(documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(documents))
	}
	if documents[0].Content == "" {
		t.Fatalf("expected extracted content")
	}
}

func TestNewDocumentSourceUnsupportedFormat(t *testing.T) {
	_, err := NewDocumentSource("xml", "/tmp/input.xml", "")
	if err == nil {
		t.Fatalf("expected unsupported format error")
	}
}

func TestNewDocumentSourceInferFromExtension(t *testing.T) {
	source, err := NewDocumentSource("auto", "/tmp/file.json", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source == nil {
		t.Fatalf("expected source instance")
	}
}

func TestDirectorySourceReadAll(t *testing.T) {
	tempDir := t.TempDir()
	_ = writeTempFile(t, tempDir, "a.json", `[{"id": 1, "content": "hello json"}]`)
	_ = writeTempFile(t, tempDir, "b.csv", "id,content\n2,hello csv\n")
	_ = writeTempFile(t, tempDir, "c.pdf", "%PDF-1.4\nHello PDF text")

	source := NewDirectorySource(tempDir)
	documents, err := source.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if len(documents) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(documents))
	}
	if documents[0].ID != 1 || documents[1].ID != 2 || documents[2].ID != 3 {
		t.Fatalf("expected sequential ids, got: %+v", documents)
	}
}
