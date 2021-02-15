package bytebuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestEmpty(t *testing.T) {
	e := Empty()
	assert.EqualValues(t, 0, e.Length())
}

func TestNewFromString(t *testing.T) {
	const expected = "foobarbazasdf"
	testByteBufImpl(t, NewFromString(expected), expected)
}
