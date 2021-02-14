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

	iovec := make([]syscall.Iovec, len(slices))
	for i, slice := range slices {
		iovec[i] = syscall.Iovec{&slice[0], uint64(len(slice))}
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
