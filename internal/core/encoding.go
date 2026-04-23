package core

import (
	"math/bits"
)

func VByteSize(v uint64) int {
	if v == 0 {
		return 1
	}

	return (bits.Len64(v) + 6) / 7
}

func EncodeVByteAt(buf []byte, offset int, v uint64, size int) {
	switch size {
	case 1:
		buf[offset] = byte(v)
	case 2:
		buf[offset] = byte(v&0x7F) | 0x80
		buf[offset+1] = byte(v >> 7)
	case 3:
		buf[offset] = byte(v&0x7F) | 0x80
		buf[offset+1] = byte((v>>7)&0x7F) | 0x80
		buf[offset+2] = byte(v >> 14)
	case 4:
		buf[offset] = byte(v&0x7F) | 0x80
		buf[offset+1] = byte((v>>7)&0x7F) | 0x80
		buf[offset+2] = byte((v>>14)&0x7F) | 0x80
		buf[offset+3] = byte(v >> 21)
	case 5:
		buf[offset] = byte(v&0x7F) | 0x80
		buf[offset+1] = byte((v>>7)&0x7F) | 0x80
		buf[offset+2] = byte((v>>14)&0x7F) | 0x80
		buf[offset+3] = byte((v>>21)&0x7F) | 0x80
		buf[offset+4] = byte(v >> 28)
	case 6:
		buf[offset] = byte(v&0x7F) | 0x80
		buf[offset+1] = byte((v>>7)&0x7F) | 0x80
		buf[offset+2] = byte((v>>14)&0x7F) | 0x80
		buf[offset+3] = byte((v>>21)&0x7F) | 0x80
		buf[offset+4] = byte((v>>28)&0x7F) | 0x80
		buf[offset+5] = byte(v >> 35)
	case 7:
		buf[offset] = byte(v&0x7F) | 0x80
		buf[offset+1] = byte((v>>7)&0x7F) | 0x80
		buf[offset+2] = byte((v>>14)&0x7F) | 0x80
		buf[offset+3] = byte((v>>21)&0x7F) | 0x80
		buf[offset+4] = byte((v>>28)&0x7F) | 0x80
		buf[offset+5] = byte((v>>35)&0x7F) | 0x80
		buf[offset+6] = byte(v >> 42)
	case 8:
		buf[offset] = byte(v&0x7F) | 0x80
		buf[offset+1] = byte(v>>7) | 0x80
		buf[offset+2] = byte(v>>14) | 0x80
		buf[offset+3] = byte(v>>21) | 0x80
		buf[offset+4] = byte(v>>28) | 0x80
		buf[offset+5] = byte(v>>35) | 0x80
		buf[offset+6] = byte(v>>42) | 0x80
		buf[offset+7] = byte(v >> 49)
	case 9:
		buf[offset] = byte(v) | 0x80
		buf[offset+1] = byte(v>>7) | 0x80
		buf[offset+2] = byte(v>>14) | 0x80
		buf[offset+3] = byte(v>>21) | 0x80
		buf[offset+4] = byte(v>>28) | 0x80
		buf[offset+5] = byte(v>>35) | 0x80
		buf[offset+6] = byte(v>>42) | 0x80
		buf[offset+7] = byte(v>>49) | 0x80
		buf[offset+8] = byte(v >> 56)
	case 10:
		buf[offset] = byte(v) | 0x80
		buf[offset+1] = byte(v>>7) | 0x80
		buf[offset+2] = byte(v>>14) | 0x80
		buf[offset+3] = byte(v>>21) | 0x80
		buf[offset+4] = byte(v>>28) | 0x80
		buf[offset+5] = byte(v>>35) | 0x80
		buf[offset+6] = byte(v>>42) | 0x80
		buf[offset+7] = byte(v>>49) | 0x80
		buf[offset+8] = byte(v>>56) | 0x80
		buf[offset+9] = byte(v >> 63)
	}
}

func DecodeVByteAt(rawBytes []byte, offset *int) (uint64, error) {
	pos := *offset
	remaining := len(rawBytes) - pos

	if remaining <= 0 {
		return 0, ErrTruncatedSequence
	}

	// fast path: 1-3 rawBytes
	b0 := rawBytes[pos]
	if b0 < 0x80 {
		*offset = pos + 1
		return uint64(b0), nil
	}

	if remaining == 1 {
		*offset = pos + 1
		return 0, ErrTruncatedSequence
	}

	b1 := rawBytes[pos+1]
	result := uint64(b0&0x7F) | uint64(b1&0x7F)<<7
	if b1 < 0x80 {
		*offset = pos + 2
		return result, nil
	}

	if remaining == 2 {
		*offset = pos + 2
		return 0, ErrTruncatedSequence
	}

	b2 := rawBytes[pos+2]
	result |= uint64(b2&0x7F) << 14
	if b2 < 0x80 {
		*offset = pos + 3
		return result, nil
	}

	if remaining == 3 {
		*offset = pos + 3
		return 0, ErrTruncatedSequence
	}

	// slow path: 4-10 rawBytes
	pos += 3
	shift := uint64(21)
	terminated := false

	for pos < len(rawBytes) && shift < 64 {
		b := rawBytes[pos]
		pos++
		result |= uint64(b&0x7F) << shift
		if b < 0x80 {
			terminated = true
			break
		}
		shift += 7
	}

	*offset = pos
	if !terminated {
		return 0, ErrVarintOverflow
	}
	return result, nil
}
