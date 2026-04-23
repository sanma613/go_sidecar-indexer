package source

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"practice/indexing-motor/internal/ingest"
)

type JSONSource struct {
	path string
}

func NewJSONSource(path string) *JSONSource {
	return &JSONSource{path: path}
}

func (s *JSONSource) ReadAll(_ context.Context) ([]ingest.Document, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("json must be an array of objects: %w", err)
	}

	documents := make([]ingest.Document, 0, len(rows))
	for _, row := range rows {
		content, extractErr := extractTextField(row)
		if extractErr != nil {
			continue
		}

		docID := uint64(len(documents) + 1)
		if rawID, ok := row["id"]; ok {
			switch value := rawID.(type) {
			case float64:
				if value > 0 {
					docID = uint64(value)
				}
			case string:
				if parsed, parseErr := strconv.ParseUint(strings.TrimSpace(value), 10, 64); parseErr == nil && parsed > 0 {
					docID = parsed
				}
			}
		}

		documents = append(documents, ingest.Document{ID: docID, Content: content})
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("no valid documents found in json input")
	}

	return documents, nil
}
