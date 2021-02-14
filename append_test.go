package bytebuf

import (
	"fmt"
	"testing"
)

func TestCombinedBuf(t *testing.T) {
	const expected = `foobarbazasdf`

	// Test every possible splitpoint for a string
	for offset := 0; offset < len(expected); offset++ {
		t.Run(fmt.Sprintf("Split=%d", offset), func(t *testing.T) {
			buf1 := NewFromSlice([]byte(expected[:offset]))
			buf2 := NewFromSlice([]byte(expected[offset:]))

			// Note: can't use Append here since that special-cases bytes
			combined := &combinedBuf{buf1, buf2}

			testByteBufImpl(t, combined, expected)
		})
	}
}
