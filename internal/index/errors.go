package index

import "errors"

var (
	ErrInvalidHeader = errors.New("index: invalid file header or magic number")
	ErrKeyNotFound   = errors.New("index: key not found in dictionary")
	ErrCorruptIndex  = errors.New("index: corrupt index, physical size is smaller than expected")
)
