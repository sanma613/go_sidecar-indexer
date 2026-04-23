package store

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"practice/indexing-motor/internal/core"
)

type SparseCheckpoint struct {
	Key            core.InvertedMemTableKey
	DictEntryIndex uint64
	PostOffset     uint64
	FirstDocID     uint64
}

func WriteSparseIndexFromFiles(dictPath, postPath, sparsePath string, stride uint64) error {
	if stride == 0 {
		stride = 64
	}

	dictFile, err := os.Open(dictPath)
	if err != nil {
		return err
	}
	defer dictFile.Close()

	postFile, err := os.Open(postPath)
	if err != nil {
		return err
	}
	defer postFile.Close()

	dictHeader := make([]byte, core.HeaderSize)
	if _, err := io.ReadFull(dictFile, dictHeader); err != nil {
		return err
	}

	if binary.BigEndian.Uint32(dictHeader[0:4]) != core.MagicDictionary {
		return fmt.Errorf("sparse index: invalid dictionary header")
	}
	entryCount := binary.BigEndian.Uint64(dictHeader[8:16])

	checkpoints := make([]SparseCheckpoint, 0, max(int(entryCount/stride)+1, 1))
	entryBuf := make([]byte, core.DictionaryEntrySize)
	for entryIndex := uint64(0); entryIndex < entryCount; entryIndex += stride {
		entryOffset := int64(core.HeaderSize + entryIndex*core.DictionaryEntrySize)
		if _, err := dictFile.ReadAt(entryBuf, entryOffset); err != nil {
			return err
		}

		checkpoint := SparseCheckpoint{
			Key: core.InvertedMemTableKey{
				Hash1: binary.BigEndian.Uint64(entryBuf[0:8]),
				Hash2: binary.BigEndian.Uint64(entryBuf[8:16]),
			},
			DictEntryIndex: entryIndex,
			PostOffset:     binary.BigEndian.Uint64(entryBuf[16:24]),
		}

		size := binary.BigEndian.Uint32(entryBuf[24:28])
		if size > 0 {
			postingData := make([]byte, size)
			if _, err := postFile.ReadAt(postingData, int64(checkpoint.PostOffset)); err != nil {
				return err
			}
			cursor := 0
			firstDocID, err := core.DecodeVByteAt(postingData, &cursor)
			if err == nil {
				checkpoint.FirstDocID = firstDocID
			}
		}

		checkpoints = append(checkpoints, checkpoint)
	}

	return writeSparseIndexFile(sparsePath, checkpoints)
}

func writeSparseIndexFile(path string, checkpoints []SparseCheckpoint) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if err := core.WriteHeader(writer, core.MagicSparseIndex, uint64(len(checkpoints))); err != nil {
		return err
	}

	record := make([]byte, core.SparseCheckpointSize)
	for _, checkpoint := range checkpoints {
		binary.BigEndian.PutUint64(record[0:8], checkpoint.Key.Hash1)
		binary.BigEndian.PutUint64(record[8:16], checkpoint.Key.Hash2)
		binary.BigEndian.PutUint64(record[16:24], checkpoint.DictEntryIndex)
		binary.BigEndian.PutUint64(record[24:32], checkpoint.PostOffset)
		binary.BigEndian.PutUint64(record[32:40], checkpoint.FirstDocID)
		if _, err := writer.Write(record); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}
	return file.Sync()
}
