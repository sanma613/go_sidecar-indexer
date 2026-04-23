package index

func NewTokenLUT() [256]bool {
	var lut [256]bool
	delimiters := " \t\n\r\f\v" +
		",.;:!?¡¿" +
		"\"'()[]{}<>" +
		"/\\|-_=+*&%$#@^~`"

	for _, char := range delimiters {
		lut[byte(char)] = true
	}

	return lut
}

var lut = NewTokenLUT()

type TokenIterator struct {
	rawBytes     []byte
	currentToken []byte
	cursor       uint64
	err          error
}

func LowerCaseFilter(token []byte) {
	for i := range token {
		if token[i] >= 'A' && token[i] <= 'Z' {
			token[i] += 32
		}
	}
}

func NewTokenIterator(rawBytes []byte) *TokenIterator {
	return &TokenIterator{
		rawBytes: rawBytes,
		cursor:   uint64(0),
	}
}

func (it *TokenIterator) Next() bool {
	if it.err != nil || it.cursor >= uint64(len(it.rawBytes)) {
		return false
	}

	for it.cursor < uint64(len(it.rawBytes)) {
		currentByte := it.rawBytes[it.cursor]

		if !lut[currentByte] {
			break
		}

		it.cursor++
	}

	if it.cursor >= uint64(len(it.rawBytes)) {
		return false
	}

	start := it.cursor
	for it.cursor < uint64(len(it.rawBytes)) {
		currentByte := it.rawBytes[it.cursor]

		if lut[currentByte] {
			break
		}

		it.cursor++
	}

	it.currentToken = it.rawBytes[start:it.cursor]

	return true

}

func (it *TokenIterator) Value() []byte {
	return it.currentToken
}

func (it *TokenIterator) Error() error {
	return it.err
}
