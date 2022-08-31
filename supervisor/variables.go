package supervisor

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
	"github.com/thejerf/suture/v4"
)

type Variable interface {
	fmt.Stringer
}

type SrcVar struct {
	*Connector
	Name  string // unique ID in pipeline
	Spout io.Reader
	Token suture.ServiceToken
}

func (v *SrcVar) String() string {
	return fmt.Sprintf("var(%s)", v.Name)
}

type DstVar struct {
	Name   string // unique ID in pipeline
	Sink   io.WriteCloser
	waitCh chan struct{}
}

func NewSink(name string, dst io.WriteCloser) *DstVar {
	return &DstVar{Name: name, Sink: dst, waitCh: make(chan struct{})}
}

func (v *DstVar) String() string {
	return fmt.Sprintf("var(%s)", v.Name)
}

func (v *DstVar) Write(p []byte) (n int, err error) {
	return v.Sink.Write(p)
}

func (v *DstVar) Close() error {
	log.Debug().Str("var", v.Name).Msg("Closing (EOF)")
	close(v.waitCh)
	return v.Sink.Close()
}

// WaitClosed will block until this variable is closed with Close()
func (v *DstVar) WaitClosed(ctx context.Context) error {
	select {
	case <-ctx.Done():
	case <-v.waitCh:
	}
	return nil
}

type BufferSink struct {
	*bytes.Buffer
	Closed bool
}

func (bf *BufferSink) Write(p []byte) (n int, err error) {
	if bf.Closed {
		return 0, io.EOF
	}
	return bf.Buffer.Write(p)
}

func (bf *BufferSink) Close() error {
	bf.Closed = true
	return nil
}

// NewBufferSink will store any incoming data into the passed buf.
func NewBufferSink() *BufferSink {
	return &BufferSink{&bytes.Buffer{}, false}
}
