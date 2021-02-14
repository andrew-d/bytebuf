// +build linux darwin

// TODO: also freebsd/etc.?

package bytebuf

import (
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"unsafe"
)

func maybeWritev(w io.Writer, slices [][]byte) (int64, bool, error) {
	var (
		conn syscall.RawConn
		err  error
	)
	switch v := w.(type) {
	case *os.File:
		conn, err = v.SyscallConn()

	case *net.TCPConn:
		conn, err = v.SyscallConn()

	default:
		return 0, false, nil
	}

	iovec := make([]syscall.Iovec, 0, len(slices))
	for _, slice := range slices {
		if len(slice) == 0 {
			continue
		}

		iovec = append(iovec, syscall.Iovec{
			Base: &slice[0],
			Len:  uint64(len(slice)),
		})
	}

	// Don't make a syscall for a zero-length write
	if len(iovec) == 0 {
		return 0, true, nil
	}

	var (
		n     uintptr
		errno syscall.Errno
	)
	err = conn.Write(func(fd uintptr) bool {
		n, _, errno = syscall.Syscall(
			syscall.SYS_WRITEV,
			fd,
			uintptr(unsafe.Pointer(&iovec[0])),
			uintptr(len(iovec)),
		)

		// Retry if we're interrupted or would block; the conn.Write
		// function will wait for writes to be available.
		if errno == syscall.EINTR || errno == syscall.EAGAIN {
			return false
		}
		return true
	})
	if err == nil && errno != 0 {
		err = fmt.Errorf("writev failed with error: %d", errno)
	}

	return int64(n), true, err
}
