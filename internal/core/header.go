package core

import (
	"encoding/binary"
	"io"
)

func WriteHeader(w io.Writer, magic uint32, count uint64) error {

	var header [HeaderSize]byte

	binary.BigEndian.PutUint32(header[0:4], magic)
	binary.BigEndian.PutUint32(header[4:8], CurrentVersion)
	binary.BigEndian.PutUint64(header[8:16], count)

	_, err := w.Write(header[:])
	if err != nil {
		return err
	}

	return nil

}
