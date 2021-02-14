// +build !linux,!darwin

package bytebuf

import (
	"net"
	"os"
)

func maybeSendfile(dst *net.TCPConn, src *os.File, l int64) (n int64, handled bool, err error) {
	return 0, false, nil
}
