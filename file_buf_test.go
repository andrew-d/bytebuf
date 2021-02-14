package bytebuf

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileBuf(t *testing.T) {
	const expected = `foobarbaz`

	f, err := ioutil.TempFile(t.TempDir(), "")
	require.NoError(t, err)
	defer f.Close()

	_, err = f.WriteString(expected)
	require.NoError(t, err)

	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)

	buf, err := NewFromFile(f)
	if assert.NoError(t, err) {
		testByteBufImpl(t, buf, expected)
	}
}
