// +build linux

package bytebuf

import (
	"io"
	"sync/atomic"
	"syscall"

	"golang.org/x/sys/unix"
)

var copyFileRangeSupported uint32 = 1

// copy_file_range will return spurious errors if you pass a chunk size that's
// too large - let's artificially limit this to 100 MiB for now.
//
// This is a variable so we can override it in testing.
var maxCopyFileRangeSize int = 100 * 1024 * 1024

func maybeCopyFileRange(dst, src syscall.Conn, remain int64) (int64, bool, error) {
	srcConn, err := src.SyscallConn()
	if err != nil {
		return 0, false, nil
	}

	dstConn, err := dst.SyscallConn()
	if err != nil {
		return 0, false, nil
	}

	var (
		written   int64
		srcOffset int64
	)

	for remain > 0 {
		n := maxCopyFileRangeSize
		if n > int(remain) {
			n = int(remain)
		}

		var (
			err1, err2, werr error
			currWritten      int64
		)
		err1 = srcConn.Read(func(srcfd uintptr) bool {
			err2 = dstConn.Write(func(dstfd uintptr) bool {
				currWritten, werr = copyFileRange(
					int(dstfd),
					int(srcfd),
					nil, // use destination file offset
					&srcOffset,
					n,
				)

				// Always succeed here, since we'll retry the
				// top-level Read function if there's something
				// that went wrong, which then re-runs this
				// Write call.
				return true
			})

			// Update lengths unconditionally
			if currWritten > 0 {
				written += currWritten
				remain -= currWritten
			}

			// If we get an error from Write, we're done
			if err2 != nil {
				return true
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

		// Gather error
		err = nil
		if err1 != nil {
			err = err1
		} else if err2 != nil {
			err = err2
		} else if werr != nil {
			err = werr
		}

		// Handle the error
		switch err {
		case syscall.ENOSYS:
			// This syscall is not supported
			atomic.StoreUint32(&copyFileRangeSupported, 0)
			return 0, false, nil

		case syscall.EXDEV, syscall.EINVAL, syscall.EOPNOTSUPP, syscall.EPERM:
			// Something went wrong with this call to copy_file_range(2),
			// but it's supported on this kernel; just assume we did
			// nothing
			return 0, false, nil

		case nil:
			//
			// If the most recent iteration wrote 0 bytes, but we haven't
			// yet hit 0 remaining bytes, then we reached EOF and we should
			// return. We return 'io.EOF' to indicate that we wrote less
			// than what we're expecting.
			if currWritten == 0 && remain > 0 {
				return written, true, io.EOF
			}

		default:
			// Some other error
			return written, true, err
		}
	}

	// Nothing remaining!
	return written, true, nil
}

func copyFileRange(dst, src int, dstOffset, srcOffset *int64, max int) (int64, error) {
	var (
		n   int
		err error
	)
	for {
		n, err = unix.CopyFileRange(src, srcOffset, dst, dstOffset, max, 0)
		if err != syscall.EINTR {
			break
		}
	}

	return int64(n), err
}
