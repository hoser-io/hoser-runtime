package supervisor

import (
	"bytes"
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
	Name string // unique ID in pipeline
	Sink io.WriteCloser
}

func (v *DstVar) String() string {
	return fmt.Sprintf("var(%s)", v.Name)
}

func (v *DstVar) Write(p []byte) (n int, err error) {
	return v.Sink.Write(p)
}

func (v *DstVar) Close() error {
	log.Debug().Str("var", v.Name).Msg("Closing (EOF)")
	return v.Sink.Close()
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
