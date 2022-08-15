package supervisor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hoser-io/hoser-runtime/hosercmd"
	"github.com/stretchr/testify/assert"
)

type TestPipeline struct {
	Root *Supervisor
	*Pipeline
}

func NewTestPipe(t *testing.T) *TestPipeline {
	t.Helper()
	root := New(t.TempDir())
	pipeline, err := root.AddPipeline(t.Name())
	assert.NoError(t, err)
	return &TestPipeline{
		Root:     root,
		Pipeline: pipeline,
	}
}

func TestStartExitImmediately(t *testing.T) {
	p := NewTestPipe(t)
	exiter, err := p.StartProcess("exiter", "true", nil)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	errch := p.Root.ServeBackground(ctx)
	info, err := exiter.Wait(ctx, []ProcState{ProcFinished})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, nil, info.Err)
	assert.Equal(t, 0, info.Rc)
	cancel()
	<-errch
}

func TestStartFailImmediately(t *testing.T) {
	p := NewTestPipe(t)
	exiter, err := p.StartProcess("exiter", "false", nil)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	errch := p.Root.ServeBackground(ctx)
	info, err := exiter.Wait(ctx, []ProcState{ProcFinished})
	if err != nil {
		t.Fatal(err)
	}
	assert.Error(t, info.Err)
	assert.Equal(t, 1, info.Rc)
	cancel()
	<-errch
}

func TestStartNamedArgs(t *testing.T) {
	p := NewTestPipe(t)
	cat, err := p.StartProcess("cat", "cat", []hosercmd.Arg{&hosercmd.NamedArg{In: "in"}, &hosercmd.NamedArg{Out: "out"}})
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	errch := p.Root.ServeBackground(ctx)
	info, err := cat.Wait(ctx, []ProcState{ProcRunning})
	if err != nil {
		t.Fatal(err)
	}
	if assert.NotNil(t, cat) {
		assert.Len(t, cat.Ins, 2, "expected 2 in valves on cat, status: %v", info)
		assert.Len(t, cat.Outs, 2, "expected 3 out valves on cat, status: %v", info)
	}

	cancel()
	<-errch
}

func TestVariable2Variable(t *testing.T) {
	p := NewTestPipe(t)
	src, err := p.CreateSpout("src", strings.NewReader("test string"))
	assert.NoError(t, err)
	out := NewBufferSink()
	dst, err := p.CreateSink("dst", out)
	assert.NoError(t, err)
	src.SendTo(dst)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	errch := p.Root.ServeBackground(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-errch
	assert.Equal(t, "test string", out.String())
	assert.True(t, out.Closed)
}

func args(as ...string) (result []hosercmd.Arg) {
	for _, a := range as {
		result = append(result, hosercmd.StringArg(a))
	}
	return result
}

func Test3StageFilter(t *testing.T) {
	p := NewTestPipe(t)
	inVar, err := p.CreateSpout("in", strings.NewReader(`
test string
bad line
good
`))
	assert.NoError(t, err)

	filter, err := p.StartProcess("filter", "grep", args("-v", "bad"))
	assert.NoError(t, err)

	out := NewBufferSink()
	outVar, err := p.CreateSink("out", out)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	p.Root.ServeBackground(ctx)
	_, err = filter.Wait(ctx, []ProcState{ProcRunning})
	if !assert.NoError(t, err) {
		return
	}

	inVar.SendTo(p.Processes["filter"].Ins["stdin"])
	p.Processes["filter"].Outs["stdout"].SendTo(outVar)

	info, err := filter.Wait(ctx, []ProcState{ProcFinished})
	t.Logf("Info: %v", info)
	assert.Equal(t, "\ntest string\ngood\n", out.String())
}

func TestFindSink(t *testing.T) {
	p := NewTestPipe(t)
	p.CreateSink("sinkA", NewBufferSink())
	sink, err := p.FindSink("sinkA")
	assert.NotNil(t, sink)
	assert.NoError(t, err)
}

func TestFindSource(t *testing.T) {
	p := NewTestPipe(t)
	p.CreateSpout("srcA", NewBufferSink())
	src, err := p.FindSource("srcA")
	assert.NotNil(t, src)
	assert.NoError(t, err)
}

func TestFindProc(t *testing.T) {
	p := NewTestPipe(t)
	_, err := p.StartProcess("catter", "cat", nil)
	assert.NoError(t, err)

	proc := p.FindProcess("catter")
	assert.NotNil(t, proc)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	proc.Serve(ctx)

	in, err := p.FindIn("catter", "stdin")
	assert.NoError(t, err)
	assert.NotNil(t, in)

	out, err := p.FindOut("catter", "stdout")
	assert.NoError(t, err)
	assert.NotNil(t, out)
}

func TestStop(t *testing.T) {
	p := NewTestPipe(t)
	yesser, err := p.StartProcess("yesser", "yes", args())
	assert.NoError(t, err)

	header, err := p.StartProcess("header", "head", args("-n", "3"))
	assert.NoError(t, err)

	out := NewBufferSink()
	outVar, err := p.CreateSink("out", out)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errch := p.Root.ServeBackground(ctx)
	yesser.Outs["stdout"].SendTo(header.Ins["stdin"])
	header.Outs["stdout"].SendTo(outVar)

	err = p.ExitWhen(ctx, header.Name)
	assert.NoError(t, err)

	assert.Equal(t, "y\ny\ny\n", out.String())
	err = <-errch
	assert.ErrorIs(t, err, context.Canceled)
}
