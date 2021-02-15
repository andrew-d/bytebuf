// +build !linux,!darwin

package bytebuf

import (
	"syscall"
)

func maybeSendfile(dst, src syscall.Conn, l int64) (n int64, handled bool, err error) {
	return 0, false, nil
}
