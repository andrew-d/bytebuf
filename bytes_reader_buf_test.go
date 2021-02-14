package bytebuf

import (
	"bytes"
	"testing"
)

func TestBytesReaderBuf(t *testing.T) {
	const expected = `foobarbaz`
	r := bytes.NewReader([]byte(expected))
	testByteBufImpl(t, NewFromBytesReader(r), expected)
}
