package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"practice/indexing-motor/internal/core"
	"practice/indexing-motor/internal/index"
	"practice/indexing-motor/internal/ingest"
	"practice/indexing-motor/internal/store"
)

type MemTableIndexer struct {
	activeMem  *index.InvertedMemTable
	immutable  *index.InvertedMemTable
	flushDone  chan struct{}
	flushErr   error
	flushErrMu sync.Mutex
	sparseStep uint64
	threshold  int64
	hardLimit  int64
}

func NewMemTableIndexer(sparseStep uint64) *MemTableIndexer {
	return NewMemTableIndexerWithLimits(sparseStep, core.MemTableThreshold, core.MemTableHardLimit)
}

func NewMemTableIndexerWithLimits(sparseStep uint64, threshold, hardLimit int64) *MemTableIndexer {
	if sparseStep == 0 {
		sparseStep = 64
	}
	if threshold <= 0 {
		threshold = core.MemTableThreshold
	}
	if hardLimit <= 0 {
		hardLimit = core.MemTableHardLimit
	}
	if hardLimit <= threshold {
		hardLimit = threshold + 1
	}
	return &MemTableIndexer{
		activeMem:  index.NewInvertedMemTable(),
		flushDone:  make(chan struct{}, 1),
		sparseStep: sparseStep,
		threshold:  threshold,
		hardLimit:  hardLimit,
	}
}

func (m *MemTableIndexer) AddDocument(doc ingest.Document) error {
	if err := m.awaitFlushIfNeeded(); err != nil {
		return err
	}

	it := index.NewTokenIterator([]byte(doc.Content))
	seen := make(map[core.InvertedMemTableKey]struct{})
	for it.Next() {
		if err := m.awaitFlushIfNeeded(); err != nil {
			return err
		}
		if err := m.rotateIfNeeded(); err != nil {
			return err
		}

		token := append([]byte(nil), it.Value()...)
		index.LowerCaseFilter(token)

		h1, h2 := core.DoubleHash(token)
		key := core.NewInvertedMemTableKey(h1, h2)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		m.activeMem.Add(key, doc.ID)
	}

	return it.Error()
}

func (m *MemTableIndexer) FlushFinal(dictPath, postPath string) error {
	if err := m.awaitFlushIfNeeded(); err != nil {
		return err
	}

	if err := ensureOutputDirectories(dictPath, postPath); err != nil {
		return err
	}

	if err := m.activeMem.Save(dictPath, postPath); err != nil {
		return err
	}

	sparsePath := sparsePathFromDict(dictPath)
	if err := store.WriteSparseIndexFromFiles(dictPath, postPath, sparsePath, m.sparseStep); err != nil {
		return err
	}

	return nil
}

func (m *MemTableIndexer) awaitFlushIfNeeded() error {
	if m.activeMem.CurrentSize < m.hardLimit || m.immutable == nil {
		return m.getFlushError()
	}

	<-m.flushDone
	m.immutable = nil
	return m.getFlushError()
}

func (m *MemTableIndexer) rotateIfNeeded() error {
	if m.activeMem.CurrentSize < m.threshold || m.immutable != nil {
		return nil
	}

	m.immutable = m.activeMem
	m.activeMem = index.NewInvertedMemTable()

	timestamp := time.Now().UnixNano()
	dictPath := fmt.Sprintf("data/dict-%016d.sst", timestamp)
	postPath := fmt.Sprintf("data/post-%016d.sst", timestamp)
	sparsePath := sparsePathFromDict(dictPath)

	if err := ensureOutputDirectories(dictPath, postPath); err != nil {
		return err
	}

	go func(mem *index.InvertedMemTable, dictPath, postPath, sparsePath string) {
		if err := mem.Save(dictPath, postPath); err != nil {
			m.setFlushError(err)
			m.flushDone <- struct{}{}
			return
		}
		if err := store.WriteSparseIndexFromFiles(dictPath, postPath, sparsePath, m.sparseStep); err != nil {
			m.setFlushError(err)
			m.flushDone <- struct{}{}
			return
		}

		m.flushDone <- struct{}{}
	}(m.immutable, dictPath, postPath, sparsePath)

	return nil
}

func sparsePathFromDict(dictPath string) string {
	ext := filepath.Ext(dictPath)
	if ext == "" {
		return dictPath + ".sidx"
	}
	return strings.TrimSuffix(dictPath, ext) + ".sidx"
}

func (m *MemTableIndexer) setFlushError(err error) {
	m.flushErrMu.Lock()
	defer m.flushErrMu.Unlock()
	if m.flushErr == nil {
		m.flushErr = err
	}
}

func (m *MemTableIndexer) getFlushError() error {
	m.flushErrMu.Lock()
	defer m.flushErrMu.Unlock()
	return m.flushErr
}

func ensureOutputDirectories(dictPath, postPath string) error {
	dictDir := filepath.Dir(dictPath)
	if dictDir != "" && dictDir != "." {
		if err := os.MkdirAll(dictDir, 0o755); err != nil {
			return err
		}
	}

	postDir := filepath.Dir(postPath)
	if postDir != "" && postDir != "." {
		if err := os.MkdirAll(postDir, 0o755); err != nil {
			return err
		}
	}

	return nil
}
