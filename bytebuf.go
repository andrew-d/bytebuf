package bytebuf

import (
	"io"
)

// ByteBuf is an interface for a concurrency-safe read-only buffer of bytes
// that can be read in a variety of ways.
type ByteBuf interface {
	io.ReaderAt
	io.Closer

	// The io.WriterTo implementation for all ByteBufs is guaranteed to be
	// concurrency-safe - i.e. it does not modify an internal offset and
	// can be used multiple times without changing the written data.
	io.WriterTo

	// Length returns the length of this ByteBuf.
	Length() int64

	// AsReader returns an io.Reader that reads the contents of this
	// ByteBuf. The returned Reader will only be valid so long as this
	// ByteBuf has not been closed.
	AsReader() io.Reader
}
