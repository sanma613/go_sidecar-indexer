package source

import (
	"fmt"
	"path/filepath"
	"strings"

	"practice/indexing-motor/internal/ingest"
)

func NewDocumentSource(format, inputPath, inputDir string) (ingest.DocumentSourcePort, error) {
	if strings.TrimSpace(inputDir) != "" {
		return NewDirectorySource(inputDir), nil
	}

	resolvedFormat := strings.ToLower(strings.TrimSpace(format))
	if resolvedFormat == "" || resolvedFormat == "auto" {
		ext := strings.ToLower(filepath.Ext(strings.TrimSpace(inputPath)))
		switch ext {
		case ".csv":
			resolvedFormat = "csv"
		case ".json":
			resolvedFormat = "json"
		case ".pdf":
			resolvedFormat = "pdf"
		default:
			return nil, fmt.Errorf("could not infer format from extension %q", ext)
		}
	}

	switch resolvedFormat {
	case "csv":
		return NewCSVSource(inputPath), nil
	case "json":
		return NewJSONSource(inputPath), nil
	case "pdf":
		return NewPDFSource(inputPath), nil
	default:
		return nil, fmt.Errorf("unsupported format %q (expected: auto|csv|json|pdf)", resolvedFormat)
	}
}
