package tests

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoser-io/hoser-runtime/hosercmd"
	"github.com/hoser-io/hoser-runtime/interpreter"
	"github.com/hoser-io/hoser-runtime/supervisor"
	"github.com/stretchr/testify/assert"
)

func TestExamplesSucceed(t *testing.T) {
	items, err := os.ReadDir(filepath.Join("testdata", "examples"))
	assert.NoError(t, err)
	for _, item := range items {
		if item.IsDir() {
			path, err := filepath.Abs(filepath.Join("testdata", "examples", item.Name()))
			assert.NoError(t, err)
			t.Run(item.Name(), func(t *testing.T) {
				runExample(t, path)
			})
		}
	}
}

func runExample(t *testing.T, path string) {
	varDir := t.TempDir()
	origWorkingDir, _ := os.Getwd()
	assert.NoError(t, os.Chdir(varDir))
	t.Cleanup(func() { os.Chdir(origWorkingDir) })

	pipeFd, err := os.Open(filepath.Join(path, "pipe.hos"))
	if !assert.NoError(t, err) {
		return
	}

	input, err := ioutil.ReadFile(filepath.Join(path, "input.txt"))
	if err != nil {
		t.Logf("no input.txt file found (%v): not passing input to pipeline", err)
	} else {
		assert.NoError(t, ioutil.WriteFile("input.txt", input, 0644))
	}

	super := supervisor.New(t.TempDir())
	inter := interpreter.New(super)
	cmds, err := hosercmd.ReadFiles(pipeFd)
	if err != nil {
		t.Fatalf("bad pipe.hos file: %v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errch := super.ServeBackground(ctx)
	for _, cmd := range cmds {
		err := inter.Exec(ctx, cmd)
		if err != nil {
			super.Close()
			cmdJson, _ := cmd.Body.MarshalJSON()
			t.Fatalf("command '%s %s' failed: %v", cmd.Code, cmdJson, err)
			return
		}
	}
	err = <-errch
	if err == context.DeadlineExceeded {
		t.Errorf("did not stop after 10 seconds (make sure exit command is used)")
	}

	wantOutput, err := ioutil.ReadFile(filepath.Join(path, "output.txt"))
	if err != nil {
		t.Logf("no want output.txt found, ignoring output: %v", err)
	}

	gotOutput, err := ioutil.ReadFile("output.txt")
	if err != nil {
		t.Logf("no output.txt written to: %v", err)
		if len(wantOutput) > 0 {
			t.Errorf("expected output.txt from pipeline, got none: %v", err)
		}
	}

	assert.Equal(t, string(wantOutput), string(gotOutput), "output of pipe does not match output.txt in example")
}
