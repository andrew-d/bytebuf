package bytebuf

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFromReader(t *testing.T) {
	const expected = "foobarbaz"

	var b bytes.Buffer
	_, err := b.WriteString(expected)
	require.NoError(t, err)

	// Use a wrapper type to avoid any specialization.
	buf, err := NewFromReader(struct{ io.Reader }{&b}, t.TempDir())
	if assert.NoError(t, err) {
		testByteBufImpl(t, buf, expected)
	}
}

func TestReadAll(t *testing.T) {
	const expected = `foobarbaz`

	sbuf := NewFromSlices([]byte("foo"), []byte("bar"), []byte("baz"))

	f := makeTempFile(t, expected)
	defer f.Close()
	fbuf, err := NewFromFile(f)
	require.NoError(t, err)

	cbuf1 := NewFromSlice([]byte(expected[:4]))
	cbuf2 := NewFromSlice([]byte(expected[4:]))

	// Note: can't use Append here since that special-cases bytes
	combined := &combinedBuf{cbuf1, cbuf2}

	testCases := []struct {
		Name string
		Buf  ByteBuf
	}{
		{"Slice", sbuf},
		{"File", fbuf},
		{"Combined", combined},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			data, err := ReadAll(testCase.Buf)
			if assert.NoError(t, err) {
				assert.Equal(t, expected, string(data))
			}
		})
	}
}
