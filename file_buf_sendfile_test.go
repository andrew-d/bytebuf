// +build linux

package bytebuf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileBufLarge(t *testing.T) {
	oldSize := maxSendfileSize
	maxSendfileSize = 4 * 1024 * 1024
	t.Cleanup(func() {
		maxSendfileSize = oldSize
	})

	const ss = "i'm a data line\n"
	nRepeats := (maxSendfileSize / len(ss)) + 1
	largeBuf := strings.Repeat(ss, nRepeats)

	f := makeTempFile(t, largeBuf)
	defer f.Close()
	require.NoError(t, f.Sync())

	buf, err := NewFromFile(f)
	if assert.NoError(t, err) {
		assertCopyViaConn(t, buf, largeBuf)
	}
}
