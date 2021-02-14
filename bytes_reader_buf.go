package bytebuf

import (
	"bytes"
	"io"
)

// bytesReaderBuf is a ByteBuf that's backed by a bytes.Reader
type bytesReaderBuf struct {
	r *bytes.Reader
}

var _ ByteBuf = (*bytesReaderBuf)(nil)

// NewFromBytesReader creates a ByteBuf from an underlying file.
func NewFromBytesReader(r *bytes.Reader) ByteBuf {
	ret := &bytesReaderBuf{r: r}
	return ret
}

// Length implements ByteBuf
func (b *bytesReaderBuf) Length() int64 {
	return int64(b.r.Len())
}

// AsReader implements ByteBuf
func (b *bytesReaderBuf) AsReader() io.Reader {
	return io.NewSectionReader(b.r, 0, b.Length())
}

// WriteTo implements io.WriterTo
func (b *bytesReaderBuf) WriteTo(w io.Writer) (n int64, err error) {
	// NOTE: the underlying bytes.Reader has a WriteTo implementation that
	// modifies the buffer, so we can't use it. Instead, use AsReader() -
	// and we should see if we can optimize this.
	return io.Copy(w, b.AsReader())
}

// ReadAt implements io.ReaderAt
func (b *bytesReaderBuf) ReadAt(p []byte, off int64) (int, error) {
	return b.r.ReadAt(p, off)
}

func (b *bytesReaderBuf) Close() error {
	b.r = nil
	return nil
}
