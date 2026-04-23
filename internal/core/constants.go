package core

const (
	MagicDictionary  uint32 = 0x53504449
	MagicPostings    uint32 = 0x5350504F
	MagicSparseIndex uint32 = 0x53505349
)

const Salt = 2654435769

const (
	Version1       = 1
	CurrentVersion = Version1
)

const (
	HeaderSize           = 16
	DictionaryEntrySize  = 32
	SparseCheckpointSize = 40
)

const (
	CompareLess    = -1
	CompareEqual   = 0
	CompareGreater = 1
)

const PostingCheckpointIntervalDocs = 64

const (
	// Key 16 bytes + Pointer 8 + Metadata 8
	MemMapEntryOverhead = 32

	MemDocIDSize           = 8
	MemPostingListOverhead = 16
)

const (
	MemTableThreshold = 54 * 1024 * 1024
	MemTableHardLimit = 70 * 1024 * 1024
)
