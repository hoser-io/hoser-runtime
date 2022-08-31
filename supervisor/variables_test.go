package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitClosedForever(t *testing.T) {
	buf := NewBufferSink()
	dst := NewSink("test", buf)
	_, err := dst.Write([]byte("testing"))
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	dst.WaitClosed(ctx)
	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
}

func TestWaitCloses(t *testing.T) {
	buf := NewBufferSink()
	dst := NewSink("test", buf)
	_, err := dst.Write([]byte("testing"))
	dst.Close()
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	dst.WaitClosed(ctx)
	assert.NoError(t, ctx.Err())
}
