package core

import (
	"encoding/binary"
	"hash/fnv"
)

func DoubleHash(token []byte) (uint64, uint64) {
	saltBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(saltBytes, Salt)

	h1 := fnv.New64a()
	h1.Write(token)

	h2 := fnv.New64a()

	h2.Write(saltBytes)
	h2.Write(token)
	return h1.Sum64(), h2.Sum64()
}
