package bytebuf

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileBuf(t *testing.T) {
	const expected = `foobarbaz`

	f := makeTempFile(t, expected)
	defer f.Close()

	buf, err := NewFromFile(f)
	if assert.NoError(t, err) {
		testByteBufImpl(t, buf, expected)
	}
}

func TestFileBufLarge(t *testing.T) {
	oldSize := maxSendfileSize
	maxSendfileSize = 4 * 1024 * 1024
	t.Cleanup(func() {
		maxSendfileSize = oldSize
	})

	const ss = "i'm a data line\n"
	nRepeats := (maxSendfileSize / len(ss)) + 1
	largeBuf := strings.Repeat(ss, nRepeats)

	f := makeTempFile(t, largeBuf)
	defer f.Close()
	require.NoError(t, f.Sync())

	buf, err := NewFromFile(f)
	if assert.NoError(t, err) {
		assertCopyViaConn(t, buf, largeBuf)
	}
}

func TestFileBufShortWrite(t *testing.T) {
	const expected = `foobarbazasdf`

	f := makeTempFile(t, expected)
	defer f.Close()

	buf, err := NewFromFile(f)
	if !assert.NoError(t, err) {
		return
	}

	// Now, truncate the file by 4 bytes unexpectedly.
	newLen := int64(len(expected) - 4)
	if !assert.NoError(t, f.Truncate(newLen)) {
		return
	}

	// Verify that we don't loop infinitely and get an EOF error when we
	// write to a file or network connection; this verifies that our
	// copy_file_range or sendfile optimizations handle EOF correctly.
	t.Run("WriteToFile", func(t *testing.T) {
		f := makeTempFile(t, "")
		defer f.Close()

		n, err := buf.WriteTo(f)
		// It's okay to get either an io.EOF or no error here
		if err != nil && err != io.EOF {
			t.Errorf("Expected io.EOF or no error, but got: %v", err)
		}
		assert.EqualValues(t, newLen, n)

		_, err = f.Seek(0, io.SeekStart)
		if !assert.NoError(t, err) {
			return
		}

		data, err := ioutil.ReadAll(f)
		if assert.NoError(t, err) {
			assert.Equal(t, expected[:int(newLen)], string(data))
		}
	})

	t.Run("WriteToConn", func(t *testing.T) {
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

		// We should get an io.EOF error here
		if err != nil && err != io.EOF {
			t.Errorf("Expected io.EOF or no error, but got: %v", err)
		}

		// The amount of data should be correct either way
		assert.EqualValues(t, newLen, n)

		// Close conn to signal EOF
		conn.Close()
		t.Logf("Waiting for read to finish")
		<-readDone
		assert.Equal(t, expected[:int(newLen)], connBuf.String())
	})
}
