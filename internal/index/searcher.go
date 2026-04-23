package index

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"practice/indexing-motor/internal/core"
	"strings"

	"golang.org/x/exp/mmap"
)

type Searcher struct {
	dictReader *mmap.ReaderAt
	dictSize   uint64
	postFile   *os.File
	sparse     []sparseCheckpoint
}

func NewSearcher(dictPath, postPath string) (*Searcher, error) {
	d, err := mmap.Open(dictPath)
	if err != nil {
		return nil, err
	}

	p, err := os.Open(postPath)
	if err != nil {
		return nil, err
	}

	var header [core.HeaderSize]byte
	d.ReadAt(header[:], 0)

	magic := binary.BigEndian.Uint32(header[0:4])
	if magic != core.MagicDictionary {
		return nil, ErrInvalidHeader
	}

	count := binary.BigEndian.Uint64(header[8:16])

	sparsePath := strings.TrimSuffix(dictPath, filepath.Ext(dictPath)) + ".sidx"
	sparse, _ := loadSparseIndex(sparsePath)

	return &Searcher{
		dictReader: d,
		dictSize:   count,
		postFile:   p,
		sparse:     sparse,
	}, nil
}

func (s *Searcher) Search(key core.InvertedMemTableKey) (*PostingIterator, error) {
	if s.dictSize == 0 {
		return nil, ErrKeyNotFound
	}

	left := uint64(0)
	right := s.dictSize - 1
	if len(s.sparse) > 0 {
		left, right = sparseBounds(s.sparse, key, s.dictSize)
	}

	var buf [core.DictionaryEntrySize]byte
	var other core.InvertedMemTableKey

	for left <= right {
		mid := left + (right-left)/2
		pos := int64(core.HeaderSize + (mid * core.DictionaryEntrySize))

		if pos+core.DictionaryEntrySize > int64(s.dictReader.Len()) {
			return nil, ErrCorruptIndex
		}

		_, err := s.dictReader.ReadAt(buf[:], pos)

		if err != nil {
			return nil, err
		}

		other.Hash1 = binary.BigEndian.Uint64(buf[0:8])
		other.Hash2 = binary.BigEndian.Uint64(buf[8:16])

		r := key.Compare(other)

		switch r {
		case core.CompareEqual:
			offset := binary.BigEndian.Uint64(buf[16:24])
			size := binary.BigEndian.Uint32(buf[24:28])

			rawBytes := make([]byte, size)
			_, err := s.postFile.ReadAt(rawBytes, int64(offset))

			if err != nil {
				return nil, err
			}

			return NewPostingIterator(rawBytes), nil

		case core.CompareLess:
			if mid == 0 {
				return nil, ErrKeyNotFound
			}
			right = mid - 1

		case core.CompareGreater:
			left = mid + 1
		}

	}

	return nil, ErrKeyNotFound
}
