package supervisor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/rs/zerolog/log"
)

const namedPipesDir = "namedpipes"

const (
	// Standard valve names
	StdinValve  = "stdin"
	StderrValve = "stderr"
	StdoutValve = "stdout"
)

type Valve interface {
	fmt.Stringer
	io.Closer

	// Path returns the filepath to the named pipe
	Path() string
}

// InValve is what is passed to named args, like cat {namedarg}. Since the process itself
// will be opening the path, we only open a writer to it so that we don't block the process
// from opening it.
type InValve struct {
	PortName string
	FifoPath string

	wWaiter *fifoWaiter // If in valve is an argv parameter, wWaiter notifies Write when it is ready (W is nil)
	w       *os.File    // W is connected to the processes' port

	stdin *os.File // if stdin is called to pass to stdin, this will be saved to call Close on it
}

func (iv *InValve) String() string {
	return iv.PortName
}

func (iv *InValve) Open(ctx context.Context) {
	if iv.w != nil {
		panic("in valve is already opened")
	}
	iv.wWaiter = waitForFifo(ctx, iv.FifoPath, os.O_WRONLY)
}

// OpenStdin opens a file handle to pass immediately to a process as stdin.
// The file will be closed when the valve is closed.
func (iv *InValve) OpenStdin() (*os.File, error) {
	fd, err := os.OpenFile(iv.FifoPath, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	iv.stdin = fd
	return iv.stdin, nil
}

func (iv *InValve) Close() error {
	if iv.stdin != nil {
		iv.stdin.Close()
	}
	if iv.w != nil {
		log.Debug().Str("valve", iv.PortName).Msg("closing")
		iv.w.Close()
		iv.w = nil
		iv.wWaiter = nil
	}
	return nil
}
func (iv *InValve) Write(p []byte) (n int, err error) {
	if iv.w == nil {
		if iv.wWaiter == nil {
			return 0, fmt.Errorf("must call Open before Write")
		}

		iv.w, err = iv.wWaiter.Wait()
		if err != nil {
			return 0, fmt.Errorf("cannot open named pipe: %w", err)
		}
	}
	return iv.w.Write(p)
}

func (iv *InValve) Path() string {
	return iv.FifoPath
}

// AddInValve creates an in valve for an argv parameter, that is passed as part of argv
// to the process. Since the process will call open() on the passed in path, we can't open
// the named pipe with O_RDONLY because it will cause the open() call in the process to block,
// preventing the process from terminating in some cases.
func (p *Process) AddInValve(name string) (*InValve, error) {
	fifo, err := p.newNamedPipe(name)
	if err != nil {
		return nil, err
	}

	v := &InValve{
		PortName: name,
		FifoPath: fifo,
	}
	p.Ins[name] = v
	return v, nil
}

// OutValve is an outgoing stream of data from process. Just like InValve, it can
// be either opened for stdio, so that we open both the read and write end and give
// the read end to stdout/stderr, or we can open it for argv passing, which means
// we only open the read end and let the write end be opened by the process itself
// using open().
type OutValve struct {
	*Connector
	PortName string
	FifoPath string

	rWaiter *fifoWaiter // r will be nil, until we the other end of the FIFO is opened by someone else.
	r       *os.File    // Reading from R will read data from this processes port

	stdout *os.File // If OpenStdout is called, to close once this valve is Close()
}

func (ov *OutValve) String() string {
	return ov.PortName
}

func (ov *OutValve) Close() error {
	if ov.stdout != nil {
		ov.stdout.Close()
	}
	if ov.r != nil {
		err := ov.r.Close()
		ov.r = nil
		ov.rWaiter = nil
		return err
	}
	return nil
}

func (ov *OutValve) Read(p []byte) (n int, err error) {
	if ov.r == nil {
		if ov.rWaiter == nil {
			return 0, fmt.Errorf("must call Open before Read")
		}

		ov.r, err = ov.rWaiter.Wait()
		if err != nil {
			return 0, fmt.Errorf("open named pipe for rdonly: %w", err)
		}
	}
	return ov.r.Read(p)
}

func (ov *OutValve) Path() string {
	return ov.FifoPath
}

func (ov *OutValve) Open(ctx context.Context) error {
	ov.r = nil
	ov.rWaiter = waitForFifo(ctx, ov.FifoPath, os.O_RDONLY)
	return nil
}

// OpenStdin opens a file handle to pass immediately to a process as stdin.
// The file will be closed when the valve is closed.
func (ov *OutValve) OpenStdout() (*os.File, error) {
	fd, err := os.OpenFile(ov.FifoPath, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	ov.stdout = fd
	return ov.stdout, nil
}

// AddOutValve creates a new named pipe and opens it in rdonly mode to write
// to the process that will open up the named pipe using open() itself.
func (p *Process) AddOutValve(name string) (*OutValve, error) {
	fifo, err := p.newNamedPipe(name)
	if err != nil {
		return nil, err
	}

	v := &OutValve{
		PortName:  name,
		Connector: NewConnector(),
		FifoPath:  fifo,
	}
	v.ReadFrom(v)
	p.Outs[name] = v
	return v, nil
}

type fifoEvent struct {
	readyFifo *os.File
	err       error
}

// fifoWaiter is a separate goroutine that blocks on OpenFile and will send a message to a waiting
// channel once the open succeeds. fifoWaiter is needed because just trying to open the FIFO directly
// will cause the main goroutine to block which prevent other initialization and cancellation.
type fifoWaiter struct {
	result  *fifoEvent
	ctx     context.Context
	readyCh chan fifoEvent
}

func waitForFifo(ctx context.Context, fifo string, flags int) *fifoWaiter {
	waiter := &fifoWaiter{readyCh: make(chan fifoEvent), ctx: ctx}
	go func() {
		fd, err := os.OpenFile(fifo, flags, os.ModeNamedPipe)
		waiter.readyCh <- fifoEvent{fd, err}
	}()
	return waiter
}

func (w *fifoWaiter) Wait() (*os.File, error) {
	if w.result != nil {
		return w.result.readyFifo, w.result.err
	}
	select {
	case event := <-w.readyCh:
		w.result = &event
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	}
	return w.result.readyFifo, w.result.err
}

func (p *Process) newNamedPipe(port string) (fifo string, err error) {
	err = os.MkdirAll(filepath.Join(p.DataDir, namedPipesDir), 0755)
	if err != nil {
		return
	}
	fifo = filepath.Join(p.DataDir, namedPipesDir, port)
	err = syscall.Mkfifo(fifo, 0644)
	if err != nil {
		return "", err
	}
	return
}
