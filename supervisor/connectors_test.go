package supervisor

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnectorTransfers(t *testing.T) {
	r := strings.NewReader("test here")
	w := NewBufferSink()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	conn := NewConnector()
	conn.ReadFrom(r)
	errch := make(chan error)
	go func() {
		errch <- conn.Serve(ctx)
	}()
	conn.SendTo(w)
	err := <-errch
	assert.Error(t, err, io.EOF)
	assert.Equal(t, "test here", w.String())
}

func TestConnectorWaiting(t *testing.T) {
	r := strings.NewReader("test here")
	buf := NewBufferSink()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	conn := NewConnector()
	conn.ReadFrom(r)
	errch := make(chan error)
	go func() {
		errch <- conn.Serve(ctx)
	}()
	time.Sleep(10 * time.Millisecond)
	conn.SendTo(buf)
	err := <-errch
	assert.Error(t, err, io.EOF)
	assert.Equal(t, "test here", buf.String())
}

func TestConnectorWaitingTimeout(t *testing.T) {
	r := strings.NewReader("test here")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	conn := NewConnector()
	conn.ReadFrom(r)
	errch := make(chan error)
	go func() {
		errch <- conn.Serve(ctx)
	}()
	err := <-errch
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestConnectorDstEOF(t *testing.T) {
	r := strings.NewReader("test here")
	w := newWriter(func(buf []byte) (int, error) {
		return 0, io.EOF
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	conn := NewConnector()
	conn.ReadFrom(r)
	errch := make(chan error)
	go func() {
		errch <- conn.Serve(ctx)
	}()
	conn.SendTo(w)
	err := <-errch
	assert.ErrorIs(t, err, io.EOF)
}

type TestReader struct {
	Cb func(buf []byte) (int, error)
}

func newReader(cb func(buf []byte) (int, error)) *TestReader {
	return &TestReader{Cb: cb}
}

func (tr *TestReader) Read(buf []byte) (int, error) {
	return tr.Cb(buf)
}

type TestWriter struct {
	Cb func(buf []byte) (int, error)
}

func newWriter(cb func(buf []byte) (int, error)) *TestWriter {
	return &TestWriter{Cb: cb}
}

func (tr *TestWriter) Write(buf []byte) (int, error) {
	return tr.Cb(buf)
}

func (tr *TestWriter) Close() error {
	return nil
}
