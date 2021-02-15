package bytebuf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testByteBufImpl(t *testing.T, impl ByteBuf, expected string) {
	defer impl.Close()

	t.Run("Length", func(t *testing.T) {
		assert.EqualValues(t, len(expected), impl.Length())
	})

	t.Run("AsReader", func(t *testing.T) {
		data, err := ioutil.ReadAll(impl.AsReader())
		if assert.NoError(t, err) {
			assert.Equal(t, expected, string(data))
		}
	})

	t.Run("ReadAt", func(t *testing.T) {
		assertReadAt := func(t *testing.T, offset, length int, data string) {
			buf := make([]byte, length)
			n, err := impl.ReadAt(buf, int64(offset))
			if assert.NoError(t, err) {
				assert.Equal(t, length, n)
				assert.Equal(t, data, string(buf))
			}
		}

		// Test every possible length and offset
		for offset := 0; offset < len(expected); offset++ {
			offset := offset
			t.Run(fmt.Sprintf("Offset=%d", offset), func(t *testing.T) {
				for length := 1; length < len(expected)-offset; length++ {
					assertReadAt(t, offset, length, expected[offset:offset+length])
				}
			})
		}
	})

	t.Run("WriteTo", func(t *testing.T) {
		// NOTE: running a WriteTo test multiple times also verifies
		// that it doesn't mutate the underlying buffer, offset, etc.

		t.Run("NonVectored", func(t *testing.T) {
			var buf bytes.Buffer

			n, err := impl.WriteTo(&buf)
			require.NoError(t, err)
			require.Equal(t, impl.Length(), n)

			assert.Equal(t, expected, buf.String())
		})

		t.Run("WriteToFile", func(t *testing.T) {
			f := makeTempFile(t, "")
			defer f.Close()

			// Test writing to a file, which should exercise
			// vectored I/O or other fast paths, if supported.
			n, err := impl.WriteTo(f)
			require.NoError(t, err)
			require.Equal(t, impl.Length(), n)

			_, err = f.Seek(0, io.SeekStart)
			require.NoError(t, err)

			data, err := ioutil.ReadAll(f)
			if assert.NoError(t, err) {
				assert.Equal(t, expected, string(data))
			}
		})

		t.Run("WriteToConn", func(t *testing.T) {
			assertCopyViaConn(t, impl, expected)
		})
	})
}

// assertCopyViaConn will copy the given buffer to a net.Conn and assert that
// the data matches the expected value.
func assertCopyViaConn(t *testing.T, buf io.WriterTo, expected string) {
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer l.Close()

	var (
		connBuf  bytes.Buffer
		readDone = make(chan struct{})
	)
	go func() {
		defer close(readDone)
		conn, err := l.Accept()
		if !assert.NoError(t, err) {
			return
		}
		defer conn.Close()

		_, err = io.Copy(&connBuf, conn)
		if !assert.NoError(t, err) {
			return
		}
	}()

	// Open a connection
	t.Logf("Dialing connection to: %+v", l.Addr())
	conn, err := net.Dial(l.Addr().Network(), l.Addr().String())
	if !assert.NoError(t, err) {
		return
	}
	defer conn.Close()

	t.Logf("Writing data to connection")
	n, err := buf.WriteTo(conn)
	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, len(expected), n)

	// Close conn to signal EOF
	conn.Close()
	t.Logf("Waiting for read to finish")
	<-readDone
	assert.Equal(t, expected, connBuf.String())
}

// makeTempFile creates a temporary file with the provided data in the tests's
// TempDir.
func makeTempFile(t *testing.T, expected string) *os.File {
	f, err := ioutil.TempFile(t.TempDir(), "")
	require.NoError(t, err)

	if expected != "" {
		_, err = f.WriteString(expected)
		require.NoError(t, err)

		_, err = f.Seek(0, io.SeekStart)
		require.NoError(t, err)
	}

	return f
}
