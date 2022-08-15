package supervisor

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOpenInArgv(t *testing.T) {
	proc, err := NewProcess("test", "/usr/bin/cat", ProcessConfig{
		PrivateDir: t.TempDir(),
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	in, err := proc.AddInValve("test_valve")
	assert.NoError(t, err)
	assert.Equal(t, "test_valve", in.PortName)

	wait := make(chan error)
	go func() {
		fd, err := os.OpenFile(in.Path(), os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			wait <- err
			return
		}

		bytes, err := ioutil.ReadAll(fd)
		if err != nil {
			wait <- err
			return
		}
		assert.Equal(t, "opened", string(bytes))
		wait <- nil
	}()

	in.Open(ctx)
	fmt.Fprintf(in, "opened")
	in.Close()
	err = <-wait
	cancel()
	assert.NoError(t, err)
}

func TestOpenOutArgv(t *testing.T) {
	proc, err := NewProcess("test", "/usr/bin/cat", ProcessConfig{
		PrivateDir: t.TempDir(),
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	out, err := proc.AddOutValve("test_valve")
	assert.NoError(t, err)
	assert.Equal(t, "test_valve", out.PortName)
	go func() {
		t.Logf("opening %s", out.Path())
		fd, err := os.OpenFile(out.Path(), os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			t.Logf("error: %v", err)
			return
		}

		t.Logf("writing opened")
		fmt.Fprintf(fd, "opened")
		fd.Close()
	}()

	out.Open(ctx)
	data, err := ioutil.ReadAll(out)
	assert.NoError(t, err)
	assert.Equal(t, "opened", string(data))
	cancel()
	assert.NoError(t, err)
}

func TestForgottenValve(t *testing.T) {
	proc, err := NewProcess("test", "/usr/bin/cat", ProcessConfig{
		PrivateDir: t.TempDir(),
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	in, err := proc.AddInValve("test_valve")
	assert.NoError(t, err)
	assert.Equal(t, "test_valve", in.PortName)

	_, err = in.Write([]byte{0x00})
	assert.Error(t, err)

	in.Open(ctx)
	_, err = in.Write([]byte{0x00})
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	cancel()
}
