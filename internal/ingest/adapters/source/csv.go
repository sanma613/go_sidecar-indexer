package source

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"practice/indexing-motor/internal/ingest"
)

type CSVSource struct {
	path string
}

func NewCSVSource(path string) *CSVSource {
	return &CSVSource{path: path}
}

func (s *CSVSource) ReadAll(_ context.Context) ([]ingest.Document, error) {
	file, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("csv input is empty")
	}

	headers := records[0]
	contentIdx := -1
	idIdx := -1
	for i, header := range headers {
		normalized := strings.ToLower(strings.TrimSpace(header))
		switch normalized {
		case "content", "text", "body", "message":
			if contentIdx == -1 {
				contentIdx = i
			}
		case "id", "docid", "document_id":
			if idIdx == -1 {
				idIdx = i
			}
		}
	}
	if contentIdx == -1 {
		contentIdx = 0
	}

	documents := make([]ingest.Document, 0, max(len(records)-1, 0))
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) == 0 || contentIdx >= len(row) {
			continue
		}

		content := strings.TrimSpace(row[contentIdx])
		if content == "" {
			continue
		}

		docID := uint64(len(documents) + 1)
		if idIdx >= 0 && idIdx < len(row) {
			if parsed, parseErr := strconv.ParseUint(strings.TrimSpace(row[idIdx]), 10, 64); parseErr == nil && parsed > 0 {
				docID = parsed
			}
		}

		documents = append(documents, ingest.Document{ID: docID, Content: content})
	}

	return documents, nil
}
