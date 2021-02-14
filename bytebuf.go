package bytebuf

import (
	"io"
)

// ByteBuf is an interface for a concurrency-safe read-only buffer of bytes
// that can be read in a variety of ways.
type ByteBuf interface {
	io.ReaderAt
	io.WriterTo
	io.Closer

	// Length returns the length of this ByteBuf.
	Length() int64

	// AsReader returns an io.Reader that reads the contents of this
	// ByteBuf. The returned Reader will only be valid so long as this
	// ByteBuf has not been closed.
	AsReader() io.Reader
}
