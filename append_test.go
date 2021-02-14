package bytebuf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombinedBuf(t *testing.T) {
	t.Run("ReadAt", func(t *testing.T) {
		// Test every possible length and offset
		const expected = `foobarbazasdf`
		for offset := 0; offset < len(expected); offset++ {
			t.Run(fmt.Sprintf("Offset=%d", offset), func(t *testing.T) {
				buf1 := NewFromSlice([]byte(expected[:offset]))
				buf2 := NewFromSlice([]byte(expected[offset:]))

				// Note: can't use Append here since that special-cases bytes
				combined := &combinedBuf{buf1, buf2}

				assertReadAt := func(t *testing.T, offset, length int, data string) {
					buf := make([]byte, length)
					n, err := combined.ReadAt(buf, int64(offset))
					if assert.NoError(t, err) {
						assert.Equal(t, length, n)
						assert.Equal(t, data, string(buf))
					}
				}

				for length := 1; length < len(expected)-offset; length++ {
					assertReadAt(t, offset, length, expected[offset:offset+length])
				}
			})
		}
	})
}
