// +build !linux

package bytebuf

import (
	"syscall"
)

func maybeCopyFileRange(dst, src syscall.Conn, l int64) (n int64, handled bool, err error) {
	return 0, false, nil
}
