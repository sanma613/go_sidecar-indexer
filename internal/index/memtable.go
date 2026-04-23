package index

import (
	"practice/indexing-motor/internal/core"
	"practice/indexing-motor/internal/store"
	"slices"
)

type InvertedMemTable struct {
	CurrentSize int64
	data        map[core.InvertedMemTableKey]*PostingList
}

func NewInvertedMemTable() *InvertedMemTable {
	return &InvertedMemTable{
		data: make(map[core.InvertedMemTableKey]*PostingList),
	}
}

func (m *InvertedMemTable) Add(key core.InvertedMemTableKey, docID uint64) {
	pl, ok := m.data[key]
	if !ok {
		pl = NewPostingList(16)
		m.data[key] = pl

		m.CurrentSize += core.MemMapEntryOverhead + core.MemPostingListOverhead
	}
	bytesAdded := pl.Add(docID)
	m.CurrentSize += int64(bytesAdded)
}

func (m *InvertedMemTable) sortKeys() []core.Entry {
	entries := make([]core.Entry, len(m.data))
	i := 0
	for k, v := range m.data {
		entry := core.NewEntry(k, v.LastDocID, v.RawBytes)
		entries[i] = entry

		i++
	}

	slices.SortFunc(entries, func(a, b core.Entry) int {
		return a.Key.Compare(b.Key)
	})

	return entries
}

func (m *InvertedMemTable) Save(dictPath, postPath string) error {

	entries := m.sortKeys()

	err := store.WriteSSTable(dictPath, postPath, entries)
	if err != nil {
		return err
	}

	return nil

}
