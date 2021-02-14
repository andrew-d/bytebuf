package bytebuf

import (
	"testing"
)

func TestSliceBuf(t *testing.T) {
	b := &sliceBuf{
		slices: [][]byte{
			[]byte("foo"),
			[]byte("b"),
			[]byte("arbaz"),
			[]byte("asdf"),
		},
	}

	testByteBufImpl(t, b, "foobarbazasdf")
}
