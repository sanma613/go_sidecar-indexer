package index

import (
	"practice/indexing-motor/internal/core"
)

type PostingList struct {
	LastDocID  uint64
	DocIDCount uint8
	RawBytes   []byte
}

type PostingIterator struct {
	data       []byte
	cursor     int
	currentID  uint64
	docIDCount uint16
	err        error
}

func NewPostingList(initialCap int) *PostingList {
	return &PostingList{
		LastDocID:  uint64(0),
		DocIDCount: uint8(0),
		RawBytes:   make([]byte, 0, initialCap),
	}
}

func NewPostingIterator(rawBytes []byte) *PostingIterator {

	return &PostingIterator{
		data:       rawBytes,
		cursor:     int(0),
		docIDCount: uint16(0),
	}
}

func (p *PostingList) EnsureCap(minCap int) {
	if cap(p.RawBytes) >= minCap {
		return
	}

	newCap := max(minCap, cap(p.RawBytes)*2)

	rawBytesLen := len(p.RawBytes)

	newRawBytes := make([]byte, rawBytesLen, newCap)
	copy(newRawBytes, p.RawBytes[:rawBytesLen])
	p.RawBytes = newRawBytes
}

func (p *PostingList) Grow(n int) {
	newLen := len(p.RawBytes) + n

	p.EnsureCap(newLen)
	p.RawBytes = p.RawBytes[:newLen]
}

func (p *PostingList) Add(docID uint64) int {
	if len(p.RawBytes) > 0 && docID <= p.LastDocID {
		return 0
	}

	var value uint64
	if p.DocIDCount != 0 {
		value = docID - p.LastDocID // absoluto
	} else {
		value = docID // delta
	}

	size := core.VByteSize(value)
	startWriteAt := len(p.RawBytes)
	p.Grow(size)
	core.EncodeVByteAt(p.RawBytes, startWriteAt, value, size)

	p.LastDocID = docID
	p.DocIDCount++

	if p.DocIDCount == core.PostingCheckpointIntervalDocs {
		p.DocIDCount = 0
	}

	return size
}

func (it *PostingIterator) Next() bool {
	if it.err != nil || it.cursor >= len(it.data) {
		return false
	}

	val, err := core.DecodeVByteAt(it.data, &it.cursor)
	if err != nil {
		it.err = err
		return false
	}

	if it.docIDCount != 0 {
		it.currentID += val // delta
	} else {
		it.currentID = val // absoluto

	}

	it.docIDCount++
	if it.docIDCount == core.PostingCheckpointIntervalDocs {
		it.docIDCount = 0
	}

	return true
}

func (it *PostingIterator) Value() uint64 {
	return it.currentID
}

func (it *PostingIterator) Err() error {
	return it.err
}
