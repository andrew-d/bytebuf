package bytebuf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCombinedBuf(t *testing.T) {
	const expected = `foobarbazasdf`

	// Test every possible splitpoint for a string
	for offset := 0; offset < len(expected); offset++ {
		offset := offset
		t.Run(fmt.Sprintf("Split=%d", offset), func(t *testing.T) {
			buf1 := NewFromSlice([]byte(expected[:offset]))
			buf2 := NewFromSlice([]byte(expected[offset:]))

			// Note: can't use Append here since that special-cases bytes
			combined := &combinedBuf{buf1, buf2}

			testByteBufImpl(t, combined, expected)
		})
	}
}

func TestAppend(t *testing.T) {
	t.Run("CombineTwoSlices", func(t *testing.T) {
		buf1 := NewFromSlice([]byte("foo"))
		buf2 := NewFromSlice([]byte("bar"))

		combined := Append(buf1, buf2)

		slice, ok := combined.(*sliceBuf)
		if assert.True(t, ok) {
			assert.Equal(t, [][]byte{
				[]byte("foo"),
				[]byte("bar"),
			}, slice.slices)
		}
	})

	t.Run("CombineNestedSlices", func(t *testing.T) {
		buf1 := NewFromSlice([]byte("foo"))
		buf2 := NewFromSlice([]byte("bar"))

		// NOTE: don't use Append here to avoid coalescing.
		combined := &combinedBuf{buf1, buf2}

		buf3 := NewFromSlice([]byte("baz"))
		combined2 := Append(combined, buf3)

		slice, ok := combined2.(*combinedBuf)
		if assert.True(t, ok) {
			// The second item inside the combined buf should be a
			// slice buf that has the last two slices coalesced.
			if slice2, ok := slice.two.(*sliceBuf); assert.True(t, ok) {
				assert.Equal(t, [][]byte{
					[]byte("bar"),
					[]byte("baz"),
				}, slice2.slices)
			}
		}
	})

	t.Run("CombineNonSlices", func(t *testing.T) {
		f1 := makeTempFile(t, "foo")
		defer f1.Close()

		f2 := makeTempFile(t, "bar")
		defer f2.Close()

		buf1, err := NewFromFile(f1)
		require.NoError(t, err)
		buf2, err := NewFromFile(f2)
		require.NoError(t, err)

		combined := Append(buf1, buf2)

		testByteBufImpl(t, combined, "foobar")
	})
}
