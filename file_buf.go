package bytebuf

import (
	"io"
	"net"
	"os"
)

// fileBuf is a ByteBuf that's backed by a File.
type fileBuf struct {
	f    *os.File
	size int64
}

var _ ByteBuf = (*fileBuf)(nil)

// NewFromFile creates a ByteBuf from an underlying file.
func NewFromFile(f *os.File) (ByteBuf, error) {
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}

	ret := &fileBuf{f: f, size: st.Size()}
	return ret, nil
}

// Length implements ByteBuf
func (b *fileBuf) Length() int64 {
	return b.size
}

// AsReader implements ByteBuf
func (b *fileBuf) AsReader() io.Reader {
	return io.NewSectionReader(b, 0, b.size)
}

// WriteTo implements io.WriterTo
func (b *fileBuf) WriteTo(w io.Writer) (n int64, err error) {
	var handled bool
	switch v := w.(type) {
	case *os.File:
		// Try to use copy_file_range(2) to copy directly from the file
		// to the output file.
		n, handled, err = maybeCopyFileRange(v, b.f, b.size)

	case *net.TCPConn:
		// Try to use sendfile(2) to copy data directly from the file
		// to the connection.
		n, handled, err = maybeSendfile(v, b.f, b.size)
	}
	if handled {
		return
	}

	n, err = io.Copy(w, b.AsReader())
	return
}

// ReadAt implements io.ReaderAt
func (b *fileBuf) ReadAt(p []byte, off int64) (int, error) {
	return b.f.ReadAt(p, off)
}

func (b *fileBuf) Close() error {
	return b.f.Close()
}
