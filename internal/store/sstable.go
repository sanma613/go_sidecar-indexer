package store

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"practice/indexing-motor/internal/core"
)

func WriteEntry(w io.Writer, entry core.Entry) error {
	var dictionaryEntry [32]byte

	binary.BigEndian.PutUint64(dictionaryEntry[0:8], entry.Key.Hash1)
	binary.BigEndian.PutUint64(dictionaryEntry[8:16], entry.Key.Hash2)
	binary.BigEndian.PutUint64(dictionaryEntry[16:24], entry.Offset)
	binary.BigEndian.PutUint32(dictionaryEntry[24:28], entry.Size)

	if _, err := w.Write(dictionaryEntry[:]); err != nil {
		return err
	}

	return nil
}

func WriteSSTable(dictPath, postPath string, entries []core.Entry) error {
	dFile, err := os.Create(dictPath)
	if err != nil {
		return err
	}
	defer dFile.Close()

	pFile, err := os.Create(postPath)
	if err != nil {
		return err
	}
	defer pFile.Close()

	dWriter := bufio.NewWriter(dFile)
	pWriter := bufio.NewWriter(pFile)

	err = core.WriteHeader(dWriter, core.MagicDictionary, uint64(len(entries)))
	if err != nil {
		return err
	}
	err = core.WriteHeader(pWriter, core.MagicPostings, 0)
	if err != nil {
		return err
	}

	offset := core.HeaderSize
	for i := range entries {
		rawBytes := entries[i].RawBytes
		rawBytesLen := len(rawBytes)

		entries[i].Offset = uint64(offset)
		entries[i].Size = uint32(rawBytesLen)

		if _, err := pWriter.Write(rawBytes); err != nil {
			return err
		}

		offset += rawBytesLen
	}

	for _, entry := range entries {
		if err := WriteEntry(dWriter, entry); err != nil {
			return err
		}
	}

	if err := pWriter.Flush(); err != nil {
		return err
	}
	if err := pFile.Sync(); err != nil {
		return err
	}

	if err := dWriter.Flush(); err != nil {
		return err
	}
	if err := dFile.Sync(); err != nil {
		return err
	}

	return nil

}
