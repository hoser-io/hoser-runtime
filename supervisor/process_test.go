package supervisor

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/hoser-io/hoser-runtime/hosercmd"
	"github.com/stretchr/testify/assert"
	"github.com/thejerf/suture/v4"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func TestStartInputs(t *testing.T) {
	proc, err := NewProcess("catter", must(exec.LookPath("cat")), ProcessConfig{
		Argv:       []string{"-", "$in"},
		Ports:      map[string]hosercmd.Port{"in": hosercmd.Port{Dir: hosercmd.DirIn}},
		PrivateDir: t.TempDir(),
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	sup := suture.New("root", suture.Spec{
		EventHook: func(e suture.Event) {
			t.Log(e)
		},
	})
	sup.Add(proc.SupervisorTree())
	errch := sup.ServeBackground(ctx)

	info, err := proc.Wait(ctx, []ProcState{ProcRunning})
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, proc.Ins, 2, "expected 2 in valves on cat, status: %v", info)
	assert.Len(t, proc.Outs, 1, "expected 1 out valves on cat, status: %v", info)

	wait := make(chan struct{})
	go func() {
		fmt.Fprintf(proc.Ins["stdin"], "some\nlines\n")
		proc.Ins["stdin"].Close()
		fmt.Fprintf(proc.Ins["in"], "input\n")
		proc.Ins["in"].Close()
		wait <- struct{}{}
	}()

	buf := NewBufferSink()
	proc.Outs["stdout"].SendTo(buf)
	<-wait

	info, err = proc.Wait(ctx, []ProcState{ProcFinished})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Info: %v", info)
	assert.NoError(t, err)
	if proc != nil {
		assert.Equal(t, 0, info.Rc)
		assert.Equal(t, "some\nlines\ninput\n", buf.String())
	} else {
		t.Errorf("expected '%s' to have finished (state %s)", proc.Name, info.State)
		t.Logf("Buffer contents: %s", buf.String())
	}

	cancel()
	err = <-errch
	assert.ErrorIs(t, err, context.Canceled)
}
