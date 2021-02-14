package bytebuf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testByteBufImpl(t *testing.T, impl ByteBuf, expected string) {
	defer impl.Close()

	t.Run("Length", func(t *testing.T) {
		assert.EqualValues(t, len(expected), impl.Length())
	})

	t.Run("AsReader", func(t *testing.T) {
		data, err := ioutil.ReadAll(impl.AsReader())
		if assert.NoError(t, err) {
			assert.Equal(t, expected, string(data))
		}
	})

	t.Run("ReadAt", func(t *testing.T) {
		assertReadAt := func(t *testing.T, offset, length int, data string) {
			buf := make([]byte, length)
			n, err := impl.ReadAt(buf, int64(offset))
			if assert.NoError(t, err) {
				assert.Equal(t, length, n)
				assert.Equal(t, data, string(buf))
			}
		}

		// Test every possible length and offset
		for offset := 0; offset < len(expected); offset++ {
			t.Run(fmt.Sprintf("Offset=%d", offset), func(t *testing.T) {
				for length := 1; length < len(expected)-offset; length++ {
					assertReadAt(t, offset, length, expected[offset:offset+length])
				}
			})
		}
	})

	t.Run("WriteTo", func(t *testing.T) {
		t.Run("NonVectored", func(t *testing.T) {
			var buf bytes.Buffer

			n, err := impl.WriteTo(&buf)
			require.NoError(t, err)
			require.Equal(t, impl.Length(), n)

			assert.Equal(t, expected, string(buf.Bytes()))
		})

		t.Run("Vectored", func(t *testing.T) {
			f, err := ioutil.TempFile(t.TempDir(), "")
			require.NoError(t, err)
			defer f.Close()

			// Test writing to a file, which should exercise
			// vectored I/O or other fast paths, if supported.
			n, err := impl.WriteTo(f)
			require.NoError(t, err)
			require.Equal(t, impl.Length(), n)

			_, err = f.Seek(0, io.SeekStart)
			require.NoError(t, err)

			data, err := ioutil.ReadAll(f)
			if assert.NoError(t, err) {
				assert.Equal(t, expected, string(data))
			}
		})
	})
}
