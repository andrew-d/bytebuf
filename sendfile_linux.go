// +build linux

package bytebuf

import (
	"io"
	"syscall"
)

func maybeSendfile(dst, src syscall.Conn, l int64) (int64, bool, error) {
	fConn, err := src.SyscallConn()
	if err != nil {
		return 0, false, nil
	}

	netConn, err := dst.SyscallConn()
	if err != nil {
		return 0, false, nil
	}

	// Use the RawConns to get a FD for reading and writing. Similar to the
	// Go runtime (in src/net/sendfile_linux.go), we don't retry reads from
	// the source file.
	var (
		n    int64
		werr error
	)
	err = fConn.Read(func(fd uintptr) bool {
		n, werr = sendfileFd(netConn, fd, l)
		return true
	})

	// If we get ENOSYS or EINVAL from sendfile(2), then the kernel doesn't
	// support the syscall, or the fd is in the wrong state; we didn't move
	// any data.
	if werr == syscall.ENOSYS || werr == syscall.EINVAL {
		return 0, false, nil
	}

	// Return either error, if we got one.
	if err == nil {
		err = werr
	}

	return n, true, err
}

// sendfile(2) will only send at most 0x7ffff000 bytes in one chunk,
// but we limit things to a smaller size to prevent large transfers
// from blocking too long.
//
// This is a variable so we can override it in testing.
var maxSendfileSize int = 4 * 1024 * 1024

func sendfileFd(dst syscall.RawConn, src uintptr, remain int64) (int64, error) {
	var (
		offset  int64
		written int64
		err     error
	)
	for remain > 0 {
		n := maxSendfileSize
		if n > int(remain) {
			n = int(remain)
		}

		var (
			werr        error
			currWritten int
		)
		err = dst.Write(func(fd uintptr) bool {
			currWritten, werr = syscall.Sendfile(
				int(fd),
				int(src),
				&offset,
				n,
			)

			// Update lengths unconditionally
			if currWritten > 0 {
				written += int64(currWritten)
				remain -= int64(currWritten)
			}

			// 0-sized write but no error indicates an EOF; stop
			// here.
			if currWritten == 0 && werr == nil {
				return true
			}

			// If this is an EINTR or EAGAIN, then we return false
			// to signal that we should retry this function.
			if werr == syscall.EINTR || werr == syscall.EAGAIN {
				return false
			}

			// This is some other error; return true to stop
			// iterating.
			return true
		})
		if err == nil {
			err = werr
		}

		// If we have an error, then stop and return what we have.
		if err != nil {
			break
		}

		// If the most recent iteration wrote 0 bytes, but we haven't
		// yet hit 0 remaining bytes, then we reached EOF and we should
		// return. We return 'io.EOF' to indicate that we wrote less
		// than what we're expecting.
		if currWritten == 0 && remain > 0 {
			return written, io.EOF
		}
	}

	return written, err
}
