package core

// =============================== InvertedMemTableKey ===============================

type InvertedMemTableKey struct {
	Hash1 uint64
	Hash2 uint64
}

func NewInvertedMemTableKey(hash1, hash2 uint64) InvertedMemTableKey {
	return InvertedMemTableKey{
		Hash1: hash1,
		Hash2: hash2,
	}
}

func (k InvertedMemTableKey) Compare(other InvertedMemTableKey) int {

	if k.Hash1 < other.Hash1 {
		return CompareLess
	}
	if k.Hash1 > other.Hash1 {
		return CompareGreater
	}

	if k.Hash2 < other.Hash2 {
		return CompareLess
	}
	if k.Hash2 > other.Hash2 {
		return CompareGreater
	}

	return CompareEqual
}

// =============================== Entry ===============================

type Entry struct {
	Key       InvertedMemTableKey
	LastDocID uint64
	RawBytes  []byte
	Offset    uint64
	Size      uint32
}

func NewEntry(key InvertedMemTableKey, lastDocID uint64, rawBytes []byte) Entry {

	return Entry{
		Key:       key,
		LastDocID: lastDocID,
		RawBytes:  rawBytes,
	}
}
