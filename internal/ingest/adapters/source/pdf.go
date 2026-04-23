package source

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"

	"practice/indexing-motor/internal/ingest"
)

type PDFSource struct {
	path string
}

func NewPDFSource(path string) *PDFSource {
	return &PDFSource{path: path}
}

func (s *PDFSource) ReadAll(_ context.Context) ([]ingest.Document, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("pdf input is empty")
	}

	content := extractPrintableText(data)
	if content == "" {
		return nil, fmt.Errorf("pdf text extraction produced empty content")
	}

	return []ingest.Document{{ID: 1, Content: content}}, nil
}

func extractPrintableText(data []byte) string {
	runes := make([]rune, 0, len(data))
	for _, b := range data {
		r := rune(b)
		if r == '\n' || r == '\t' || r == ' ' {
			runes = append(runes, ' ')
			continue
		}
		if unicode.IsPrint(r) {
			runes = append(runes, r)
		}
	}

	text := strings.Join(strings.Fields(string(runes)), " ")
	return strings.TrimSpace(text)
}
