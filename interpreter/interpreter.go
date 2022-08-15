package interpreter

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hoser-io/hoser-runtime/hosercmd"
	"github.com/hoser-io/hoser-runtime/supervisor"
)

// interpreter takes hosercmd and executes them on a pipeline.

const startupWait = 5 * time.Second // wait 5 seconds for a new process to start before timing out

type Interpreter struct {
	Target *supervisor.Supervisor
}

func New(target *supervisor.Supervisor) *Interpreter {
	return &Interpreter{target}
}

func (i *Interpreter) Exec(ctx context.Context, cmd hosercmd.Command) error {
	switch b := cmd.Body.(type) {
	case *hosercmd.StartBody:
		id, err := hosercmd.ParseId(b.Id)
		if err != nil {
			return err
		}

		pipeline, ok := i.Target.Pipelines[id.Pipeline]
		if !ok {
			return errMissingPipeline(id.Pipeline)
		}
		proc, err := pipeline.StartProcess(id.Node, b.ExeFile, b.Argv)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, 5*startupWait)
		defer cancel()
		_, err = proc.Wait(ctx, []supervisor.ProcState{supervisor.ProcRunning})
		return err
	case *hosercmd.PipelineBody:
		_, err := i.Target.AddPipeline(b.Id)
		return err
	case *hosercmd.ExitBody:
		id, err := hosercmd.ParseId(b.When)
		if err != nil {
			return err
		}

		pipeline, ok := i.Target.Pipelines[id.Pipeline]
		if !ok {
			return errMissingPipeline(id.Pipeline)
		}
		return pipeline.ExitWhen(ctx, id.Node)
	case *hosercmd.SetBody:
		id, err := hosercmd.ParseId(b.Id)
		if err != nil {
			return err
		}

		pipeline, ok := i.Target.Pipelines[id.Pipeline]
		if !ok {
			return errMissingPipeline(id.Pipeline)
		}

		spout, ok := pipeline.Spouts[id.Node]
		if ok {
			spout.Spout, err = parseSpoutValue(b)
			if err != nil {
				return fmt.Errorf("bad set value: %w", err)
			}
		}

		sink, ok := pipeline.Sinks[id.Node]
		if ok {
			sink.Sink, err = parseSinkValue(b)
			if err != nil {
				return fmt.Errorf("bad set value: %w", err)
			}
		}

		// Variable does not exist yet, let's create a new one, deciding whether it's a sink
		// or source from the value.
		if b.IsSink() {
			val, err := parseSinkValue(b)
			if err != nil {
				return fmt.Errorf("bad set value: %w", err)
			}
			_, err = pipeline.CreateSink(id.Node, val)
			if err != nil {
				return err
			}
		} else {
			val, err := parseSpoutValue(b)
			if err != nil {
				return fmt.Errorf("bad set value: %w", err)
			}
			_, err = pipeline.CreateSpout(id.Node, val)
			if err != nil {
				return err
			}
		}
	case *hosercmd.PipeBody:
		srcId, err := hosercmd.ParseId(b.Src)
		if err != nil {
			return fmt.Errorf("bad src id: %w", err)
		}
		dstId, err := hosercmd.ParseId(b.Dst)
		if err != nil {
			return fmt.Errorf("bad dst id: %w", err)
		}

		srcPipeline, ok := i.Target.Pipelines[srcId.Pipeline]
		if !ok {
			return errMissingPipeline(srcId.Pipeline)
		}

		dstPipeline, ok := i.Target.Pipelines[dstId.Pipeline]
		if !ok {
			return errMissingPipeline(dstId.Pipeline)
		}

		src, err := findSrc(srcPipeline, srcId)
		if err != nil {
			return err
		}
		dst, err := findDst(dstPipeline, dstId)
		if err != nil {
			return err
		}
		src.SendTo(dst)
	default:
		return fmt.Errorf("unrecognized command: %s", cmd.Code)
	}
	return nil
}

func parseSinkValue(body *hosercmd.SetBody) (io.WriteCloser, error) {
	if u, err := url.Parse(body.Write); err == nil {
		if u.Scheme == "file" {
			return os.OpenFile(filepath.Join(u.Host, u.Path), os.O_CREATE|os.O_WRONLY, 0666)
		} else {
			return nil, fmt.Errorf("'write' URL scheme '%s' is not a recognized format for a sink", u.Scheme)
		}
	}
	return nil, fmt.Errorf("command '%s' has no recognized format for a sink", body)
}

func parseSpoutValue(body *hosercmd.SetBody) (io.Reader, error) {
	if body.Text != "" {
		return strings.NewReader(body.Text), nil
	}
	if u, err := url.Parse(body.Read); err == nil {
		if u.Scheme == "file" {
			return os.Open(filepath.Join(u.Host, u.Path))
		} else {
			return nil, fmt.Errorf("URL scheme '%s' is recognized format for a source", u.Scheme)
		}
	}
	return nil, fmt.Errorf("body '%s' has no recognized value for a source", body)
}

func findSrc(pipe *supervisor.Pipeline, id hosercmd.Ident) (supervisor.Source, error) {
	if id.Port != "" {
		return pipe.FindOut(id.Node, id.Port)
	} else {
		return pipe.FindSource(id.Node)
	}
}

func findDst(pipe *supervisor.Pipeline, id hosercmd.Ident) (io.WriteCloser, error) {
	if id.Port != "" {
		return pipe.FindIn(id.Node, id.Port)
	} else {
		return pipe.FindSink(id.Node)
	}
}

func errMissingPipeline(pipeline string) error {
	return fmt.Errorf("no pipeline named '%s'", pipeline)
}
