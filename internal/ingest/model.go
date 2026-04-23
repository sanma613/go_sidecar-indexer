package ingest

import "context"

type Document struct {
	ID      uint64
	Content string
}

type DocumentSourcePort interface {
	ReadAll(ctx context.Context) ([]Document, error)
}

type IndexerPort interface {
	AddDocument(doc Document) error
	FlushFinal(dictPath, postPath string) error
}
