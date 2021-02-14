package bytebuf

import (
	"io"
)

// Append appends one ByteBuf to another. The original buffers are unmodified.
func Append(one, two ByteBuf) ByteBuf {
	if s2, ok := two.(*sliceBuf); ok {
		// If the first buffer is also a sliceBuf, then we can return a
		// single sliceBuf that contains everything.
		if s1, ok := one.(*sliceBuf); ok {
			ret := &sliceBuf{}
			ret.slices = append(ret.slices, s1.slices...)
			ret.slices = append(ret.slices, s2.slices...)
			return ret
		}

		// If the first buffer is a combinedBuf itself, and the second
		// buffer there is a slice, then we can unwrap and combine
		// them. Essentially:
		//     (X, slice) + slice = (X, slice + slice)
		//
		// This is helpful because it allows us to avoid an unnecessary
		// layer of nesting.
		//
		// TODO: we should expand combinedBuf to handle an arbitrary
		// array of buffers.
		if s1, ok := one.(*combinedBuf); ok {
			if s12, ok := s1.two.(*sliceBuf); ok {
				ret := &sliceBuf{}
				ret.slices = append(ret.slices, s12.slices...)
				ret.slices = append(ret.slices, s2.slices...)

				return &combinedBuf{s1.one, ret}
			}
		}
	}

	return &combinedBuf{one, two}
}

type combinedBuf struct {
	one, two ByteBuf
}

func (b *combinedBuf) Length() int64 {
	return b.one.Length() + b.two.Length()
}

// AsReader implements ByteBuf
func (b *combinedBuf) AsReader() io.Reader {
	return io.MultiReader(b.one.AsReader(), b.two.AsReader())
}

// WriteTo implements io.WriterTo
func (b *combinedBuf) WriteTo(w io.Writer) (n int64, err error) {
	n1, err := b.one.WriteTo(w)
	if err != nil {
		return n1, err
	}

	n2, err := b.two.WriteTo(w)
	if err != nil {
		return n1 + n2, err
	}

	return n1 + n2, nil
}

// ReadAt implements io.ReaderAt
func (b *combinedBuf) ReadAt(p []byte, off int64) (int, error) {
	oneLen := b.one.Length()
	//twoLen := b.two.Length()

	// Handle the case where the buffer falls entirely in the first or second
	// buffer.
	endPos := off + int64(len(p))
	if off < oneLen && endPos <= oneLen {
		return b.one.ReadAt(p, off)
	}
	if off >= oneLen {
		return b.two.ReadAt(p, off-oneLen)
	}

	// If we get here, the buffer spans the two underlying buffers.
	//
	// <------- one.length -------><------- two.length ------->
	//                  off <---- len(p) ---->
	//                      ^     ^^         ^
	//                      |     ||         |
	//                      +--+--++----+----+
	//                         |        |
	//                         a        b
	//
	// As you can see from the diagram above, there's two components: a and
	// b. We can calculate the midpoint from:
	//
	//    a = one.length - off
	aLen := oneLen - off

	n1, err := b.one.ReadAt(p[:aLen], off)

	// Note: io.ReaderAt specifies that this can return io.EOF; ignore that
	// error.
	if err != nil && err != io.EOF {
		return n1, err
	}

	n2, err := b.two.ReadAt(p[aLen:], 0)
	if err != nil {
		return n1 + n2, err
	}

	return n1 + n2, nil
}

// Close implements io.Closer
func (b *combinedBuf) Close() error {
	e1 := b.one.Close()
	e2 := b.two.Close()
	if e1 != nil {
		return e1
	}
	return e2
}
