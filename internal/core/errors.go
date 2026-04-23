package core

import "errors"

var (
	ErrVarintOverflow    = errors.New("core: vbyte overflows 64-bit integer")
	ErrTruncatedSequence = errors.New("core: truncated sequence, expected more bytes")
)
