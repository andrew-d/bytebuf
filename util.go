package bytebuf

import (
	"io"
	"io/ioutil"
	"os"
)

// NewFromReader creates a ByteBuf from an io.Reader. It will buffer data to
// disk in the provided directory.
func NewFromReader(r io.Reader, dir string) (ByteBuf, error) {
	// See if this is a type that we can special-case.
	switch v := r.(type) {
	case *os.File:
		return NewFromFile(v)
	}

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
// be more efficient than calling ioutil.ReadAll(buf.AsReader()).
func ReadAll(b ByteBuf) ([]byte, error) {
	switch v := b.(type) {
	case *sliceBuf:
		ret := make([]byte, 0, int(b.Length()))
		for _, slice := range v.slices {
			ret = append(ret, slice...)
		}
		return ret, nil

	case *fileBuf:
		return ioutil.ReadAll(v.f)

	default:
		return ioutil.ReadAll(v.AsReader())
	}
}
