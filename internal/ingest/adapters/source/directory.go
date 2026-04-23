package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"practice/indexing-motor/internal/ingest"
)

type DirectorySource struct {
	dirPath string
}

func NewDirectorySource(dirPath string) *DirectorySource {
	return &DirectorySource{dirPath: dirPath}
}

func (s *DirectorySource) ReadAll(ctx context.Context) ([]ingest.Document, error) {
	entries, err := os.ReadDir(s.dirPath)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".csv" && ext != ".json" && ext != ".pdf" {
			continue
		}
		paths = append(paths, filepath.Join(s.dirPath, name))
	}

	sort.Strings(paths)
	all := make([]ingest.Document, 0, 256)
	nextID := uint64(1)
	for _, path := range paths {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		source, sourceErr := NewDocumentSource("auto", path, "")
		if sourceErr != nil {
			return nil, sourceErr
		}

		documents, readErr := source.ReadAll(ctx)
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", path, readErr)
		}

		for _, document := range documents {
			all = append(all, ingest.Document{
				ID:      nextID,
				Content: document.Content,
			})
			nextID++
		}
	}

	if len(all) == 0 {
		return nil, fmt.Errorf("no supported files found in %q", s.dirPath)
	}

	return all, nil
}
