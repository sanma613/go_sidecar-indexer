package index

import (
	"encoding/binary"
	"io"
	"os"
	"practice/indexing-motor/internal/core"
)

type sparseCheckpoint struct {
	key            core.InvertedMemTableKey
	dictEntryIndex uint64
	postOffset     uint64
	firstDocID     uint64
}

func loadSparseIndex(path string) ([]sparseCheckpoint, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	header := make([]byte, core.HeaderSize)
	if _, err := io.ReadFull(file, header); err != nil {
		return nil, err
	}
	if binary.BigEndian.Uint32(header[0:4]) != core.MagicSparseIndex {
		return nil, ErrInvalidHeader
	}

	count := binary.BigEndian.Uint64(header[8:16])
	checkpoints := make([]sparseCheckpoint, 0, count)
	record := make([]byte, core.SparseCheckpointSize)
	for i := uint64(0); i < count; i++ {
		if _, err := io.ReadFull(file, record); err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, sparseCheckpoint{
			key: core.InvertedMemTableKey{
				Hash1: binary.BigEndian.Uint64(record[0:8]),
				Hash2: binary.BigEndian.Uint64(record[8:16]),
			},
			dictEntryIndex: binary.BigEndian.Uint64(record[16:24]),
			postOffset:     binary.BigEndian.Uint64(record[24:32]),
			firstDocID:     binary.BigEndian.Uint64(record[32:40]),
		})
	}

	return checkpoints, nil
}

func sparseBounds(checkpoints []sparseCheckpoint, key core.InvertedMemTableKey, dictSize uint64) (uint64, uint64) {
	if len(checkpoints) == 0 || dictSize == 0 {
		return 0, 0
	}

	leftBlock := 0
	rightBlock := len(checkpoints) - 1
	for leftBlock <= rightBlock {
		mid := leftBlock + (rightBlock-leftBlock)/2
		comparison := checkpoints[mid].key.Compare(key)
		if comparison == core.CompareEqual {
			left := checkpoints[mid].dictEntryIndex
			right := dictSize - 1
			if mid+1 < len(checkpoints) {
				right = checkpoints[mid+1].dictEntryIndex - 1
			}
			return left, right
		}
		if comparison == core.CompareLess {
			leftBlock = mid + 1
			continue
		}
		if mid == 0 {
			return checkpoints[0].dictEntryIndex, min(dictSize-1, checkpoints[0].dictEntryIndex)
		}
		rightBlock = mid - 1
	}

	candidate := rightBlock
	if candidate < 0 {
		candidate = 0
	}

	left := checkpoints[candidate].dictEntryIndex
	right := dictSize - 1
	if candidate+1 < len(checkpoints) {
		right = checkpoints[candidate+1].dictEntryIndex - 1
	}
	if right >= dictSize {
		right = dictSize - 1
	}
	if left > right {
		left = right
	}
	return left, right
}
