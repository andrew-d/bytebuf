// +build linux

package bytebuf

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileBufCopyFileRange(t *testing.T) {
	oldSize := maxCopyFileRangeSize
	maxCopyFileRangeSize = 20 * 1024
	t.Cleanup(func() {
		maxCopyFileRangeSize = oldSize
	})

	const ss = "i'm a data line\n"
	nRepeats := (maxCopyFileRangeSize / len(ss)) + 1
	largeBuf := strings.Repeat(ss, nRepeats)

	f := makeTempFile(t, largeBuf)
	defer f.Close()
	require.NoError(t, f.Sync())

	dst := makeTempFile(t, "")
	defer dst.Close()

	buf, err := NewFromFile(f)
	if !assert.NoError(t, err) {
		return
	}

	n, err := buf.WriteTo(dst)
	assert.EqualValues(t, len(largeBuf), n)
	if !assert.NoError(t, err) {
		return
	}

	_, err = dst.Seek(0, io.SeekStart)
	if !assert.NoError(t, err) {
		return
	}

	data, err := ioutil.ReadAll(dst)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, largeBuf, string(data))
}
