// +build linux darwin

// TODO: also freebsd/etc.?

package bytebuf

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

func maybeWritev(w io.Writer, slices [][]byte) (int64, bool, error) {
	f, ok := w.(*os.File)
	if !ok {
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

	nwRaw, _, errno := syscall.Syscall(
		syscall.SYS_WRITEV,
		f.Fd(),
		uintptr(unsafe.Pointer(&iovec[0])),
		uintptr(len(iovec)),
	)
	if errno != 0 {
		return int64(nwRaw), true, fmt.Errorf("writev failed with error: %d", errno)
	}

	return int64(nwRaw), true, nil
}
