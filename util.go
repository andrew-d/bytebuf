package bytebuf

import (
	"io"
	"io/ioutil"
)

// NewFromReader creates a ByteBuf from an io.Reader. It will buffer data to
// disk in the provided directory.
func NewFromReader(r io.Reader, dir string) (ByteBuf, error) {
	f, err := ioutil.TempFile(dir, "")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		f.Close()
		return nil, err
	}

	return NewFromFile(f)
}

// ReadAll reads an entire ByteBuf into a byte slice and returns it. This may
// be faster than calling ioutil.ReadAll(buf.AsReader()).
func ReadAll(b ByteBuf) ([]byte, error) {
	switch v := b.(type) {
	case *sliceBuf:
		ret := make([]byte, int(b.Length()))
		for _, slice := range v.slices {
			n := copy(ret, slice)
			ret = ret[n:]
		}
		return ret, nil

	case *fileBuf:
		return ioutil.ReadAll(v.f)

	default:
		return ioutil.ReadAll(v.AsReader())
	}
}
