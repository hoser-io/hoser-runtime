package supervisor

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Connectors allow you take a reader and writer and connect them together using a goroutine
// that copies between them. The goal is to be able to set the destination as many times as possible
// and the connector is updated whenever a read from source finishes. If the source closes, the goroutine
// closes. If the dst closes, the dst is removed and wait for a new one.

// Waiting state - Dst == nil
// Recv new dst

type ConnectorInfo struct {
	BytesWritten int64
}

type Connector struct {
	mu   sync.Mutex
	Src  io.Reader
	Dst  io.WriteCloser
	Info ConnectorInfo

	wait chan struct{}
}

func NewConnector() *Connector {
	return &Connector{
		wait: make(chan struct{}, 1),
	}
}

func (c *Connector) IsWaiting() bool {
	return c.Dst == nil || c.Src == nil
}

func (c *Connector) Reset() {
	c.mu.Lock()
	c.Dst = nil
	c.Src = nil
	c.mu.Unlock()
}

func (c *Connector) SendTo(dst io.WriteCloser) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Dst = dst
	select {
	case c.wait <- struct{}{}: // send without blocking
	default:
	}
}

func (c *Connector) ReadFrom(src io.Reader) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Src = src
	select {
	case c.wait <- struct{}{}: // send without blocking
	default:
	}
}

// Serve will try copying from Src -> Dst. If there is no Dst (nil), then we wait blocking until
// a new Dst is received. If Dst has an EOF error or other error, the Dst is cleared. If Src has EOF,
// we exit cleanly.
func (c *Connector) Serve(ctx context.Context) (err error) {
	buf := make([]byte, 32*1024)
	defer c.Reset()
	for {
		c.mu.Lock()
		if c.IsWaiting() {
			c.mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c.wait:
			}
		} else {
			c.mu.Unlock()
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}

		c.mu.Lock()
		if c.IsWaiting() {
			c.mu.Unlock()
			continue
		}
		src := c.Src
		dst := c.Dst
		c.mu.Unlock()

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("invalid io write: %d", nw)
				}
			}
			c.Info.BytesWritten += int64(nw)
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if er == io.EOF {
			dst.Close() // signal to dst that stream is over
			return io.EOF
		} else if er != nil {
			return er
		}
	}
}
