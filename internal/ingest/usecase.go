package ingest

import "context"

type UseCase struct {
	source  DocumentSourcePort
	indexer IndexerPort
}

func NewUseCase(source DocumentSourcePort, indexer IndexerPort) *UseCase {
	return &UseCase{source: source, indexer: indexer}
}

func (uc *UseCase) Execute(ctx context.Context, dictPath, postPath string) (int, error) {
	documents, err := uc.source.ReadAll(ctx)
	if err != nil {
		return 0, err
	}

	for _, document := range documents {
		if err := uc.indexer.AddDocument(document); err != nil {
			return 0, err
		}
	}

	if err := uc.indexer.FlushFinal(dictPath, postPath); err != nil {
		return 0, err
	}

	return len(documents), nil
}
