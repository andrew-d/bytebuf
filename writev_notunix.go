// +build !linux,!darwin

package bytebuf

import (
	"io"
)

func maybeWritev(w io.Writer, slices [][]byte) (int64, bool, error) {
	return 0, false, nil
}
