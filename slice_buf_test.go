package bytebuf

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSliceBuf(t *testing.T) {
	t.Run("ReadAt", func(t *testing.T) {
		b := &sliceBuf{
			slices: [][]byte{
				[]byte("foo"),
				[]byte("b"),
				[]byte("arbaz"),
				[]byte("asdf"),
			},
		}

		assertReadAt := func(t *testing.T, offset, length int, data string) {
			buf := make([]byte, length)
			n, err := b.ReadAt(buf, int64(offset))
			if assert.NoError(t, err) {
				assert.Equal(t, length, n)
				assert.Equal(t, data, string(buf))
			}
		}

		// Test every possible length and offset
		const expected = `foobarbazasdf`
		for offset := 0; offset < len(expected); offset++ {
			t.Run(fmt.Sprintf("Offset=%d", offset), func(t *testing.T) {
				for length := 1; length < len(expected)-offset; length++ {
					assertReadAt(t, offset, length, expected[offset:offset+length])
				}
			})
		}
	})

	t.Run("WriteTo", func(t *testing.T) {
		b := &sliceBuf{
			slices: [][]byte{
				[]byte("foo"),
				[]byte("bar"),
				[]byte("baz"),
			},
		}

		t.Run("Vectored", func(t *testing.T) {
			f, err := ioutil.TempFile(t.TempDir(), "")
			require.NoError(t, err)
			defer f.Close()

			// Test writing to a file, which should exercise vectored I/O.
			n, err := b.WriteTo(f)
			require.NoError(t, err)
			require.Equal(t, b.Length(), n)

			_, err = f.Seek(0, io.SeekStart)
			require.NoError(t, err)

			data, err := ioutil.ReadAll(f)
			if assert.NoError(t, err) {
				assert.Equal(t, "foobarbaz", string(data))
			}
		})
	})
}
