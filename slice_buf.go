package bytebuf

import (
	"bytes"
	"io"
)

// sliceBuf is a ByteBuf that's backed by one or more slices of bytes.
type sliceBuf struct {
	slices      [][]byte
	singleSlice [1][]byte
}

var _ ByteBuf = (*sliceBuf)(nil)

// NewFromSlice creates a ByteBuf from an underlying slice. This avoids an
// extra allocation for the single-slice case compared to NewFromSlices.
func NewFromSlice(b []byte) ByteBuf {
	ret := &sliceBuf{}
	ret.singleSlice[0] = b
	ret.slices = ret.singleSlice[:]
	return ret
}

// NewFromSlices creates a ByteBuf from multiple slices.
func NewFromSlices(bs ...[]byte) ByteBuf {
	ret := &sliceBuf{slices: bs}
	return ret
}

// Length implements ByteBuf
func (b *sliceBuf) Length() (l int64) {
	// TODO: cache?
	for _, slice := range b.slices {
		l += int64(len(slice))
	}
	return
}

// AsReader implements ByteBuf
func (b *sliceBuf) AsReader() io.Reader {
	switch len(b.slices) {
	case 0:
		return bytes.NewReader(nil)

	case 1:
		return bytes.NewReader(b.slices[0])

	default:
		readers := make([]io.Reader, 0, len(b.slices))

		for _, v := range b.slices {
			readers = append(readers, bytes.NewReader(v))
		}

		return io.MultiReader(readers...)
	}
}

// WriteTo implements io.WriterTo
func (b *sliceBuf) WriteTo(w io.Writer) (n int64, err error) {
	n, handled, err := maybeWritev(w, b.slices)
	if handled {
		return n, err
	}

	var currN int
	for _, v := range b.slices {
		currN, err = w.Write(v)
		n += int64(currN)

		if err != nil {
			return
		}
	}
	return
}

// ReadAt implements io.ReaderAt
func (b *sliceBuf) ReadAt(p []byte, off int64) (int, error) {
	offset := int(off)
	copied := 0

	// Walk through the slices to find the index of the first slice that has
	// our data.
	sliceIdx := -1
	for i, slice := range b.slices {
		if offset < len(slice) {
			sliceIdx = i
			break
		}

		offset -= len(slice)
	}
	if sliceIdx == -1 {
		return 0, io.EOF
	}

	// Copy from the first slice
	currSlice := b.slices[sliceIdx]
	copyLen := len(p)
	if offset+copyLen > len(currSlice) {
		copyLen = len(currSlice) - offset
	}

	copied += copy(p, currSlice[offset:offset+copyLen])
	p = p[copyLen:]
	sliceIdx++

	// Now, continue copying from adjacent slices until we either run out
	// of slices or reach the end of our buffer.
	for sliceIdx < len(b.slices) && len(p) > 0 {
		slice := b.slices[sliceIdx]
		copyLen = len(p)
		if copyLen > len(slice) {
			copyLen = len(slice)
		}

		copied += copy(p, slice[:copyLen])
		p = p[copyLen:]
		sliceIdx++
	}

	if len(p) != 0 {
		return copied, io.EOF
	}
	return copied, nil
}

func (b *sliceBuf) Close() error {
	b.slices = nil
	b.singleSlice[0] = nil
	return nil
}
