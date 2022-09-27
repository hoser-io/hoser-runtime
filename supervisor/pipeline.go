package supervisor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/thejerf/suture/v4"
)

type Pipeline struct {
	*suture.Supervisor
	Creator   *Supervisor
	Name      string
	Processes map[string]*Process
	Spouts    map[string]*SrcVar
	Sinks     map[string]*DstVar
	cfg       PipelineConfig
	sid       suture.ServiceToken // pipeline's token to give to root supervisor to exit
}

type PipelineConfig struct {
	DataDir string
}

func (cfg PipelineConfig) configureDefaults() PipelineConfig {
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(os.TempDir(), fmt.Sprintf("hoser.%d", os.Getpid()))
	}
	return cfg
}

func NewPipeline(creator *Supervisor, name string, cfg PipelineConfig) *Pipeline {
	p := &Pipeline{
		Creator:   creator,
		Name:      name,
		Processes: make(map[string]*Process),
		Spouts:    make(map[string]*SrcVar),
		Sinks:     make(map[string]*DstVar),
		cfg:       cfg.configureDefaults(),
	}
	p.Supervisor = suture.New(name, suture.Spec{
		EventHook: func(e suture.Event) {
			p.handleEvent(e)
		},
	})
	return p
}

func (p *Pipeline) handleEvent(event suture.Event) {
	log.Debug().Msgf("supervisor: %v", event)
	switch e := event.(type) {
	case suture.EventServiceTerminate:
		if proc, ok := e.Service.(*Process); ok {
			if !e.Restarting && e.Err != nil {
				proc.ChangeState(func(pi *ProcInfo) {
					pi.State = ProcError
					pi.Err = e.Err.(error)
				})
			}
		}
	}
}

func (p *Pipeline) FindProcess(name string) *Process {
	return p.Processes[name]
}

func (p *Pipeline) StartProcess(name string, exe string, params *ProcessConfig) (*Process, error) {
	path, err := exec.LookPath(exe)
	if err != nil {
		return nil, err
	}

	if params == nil {
		params = &ProcessConfig{}
	}
	params.PrivateDir = filepath.Join(p.cfg.DataDir, fmt.Sprintf("process.%s", name))
	proc, err := NewProcess(name, path, *params)
	if err != nil {
		return nil, err
	}
	proc.Token = p.Add(proc.SupervisorTree())
	p.Processes[name] = proc
	return proc, nil
}

func errMissingProcess(name string) error {
	return fmt.Errorf("no process named '%s'", name)
}

func (p *Pipeline) FindIn(process, port string) (*InValve, error) {
	proc, ok := p.Processes[process]
	if !ok {
		return nil, errMissingProcess(process)
	}
	valve, ok := proc.Ins[port]
	if !ok {
		return nil, fmt.Errorf("process '%s' has no in port named '%s' (ports: %v)", process, port, proc.Ins)
	}
	return valve, nil
}

func (p *Pipeline) FindOut(process, port string) (*OutValve, error) {
	proc, ok := p.Processes[process]
	if !ok {
		return nil, errMissingProcess(process)
	}
	valve, ok := proc.Outs[port]
	if !ok {
		return nil, fmt.Errorf("process '%s' has no out port named '%s' (ports: %v)", process, port, proc.Outs)
	}
	return valve, nil
}

func (p *Pipeline) FindSource(name string) (*SrcVar, error) {
	v, ok := p.Spouts[name]
	if !ok {
		return nil, fmt.Errorf("no var found named '%s'", name)
	}
	return v, nil
}

func (p *Pipeline) FindSink(name string) (*DstVar, error) {
	v, ok := p.Sinks[name]
	if !ok {
		return nil, fmt.Errorf("no var found named '%s'", name)
	}
	return v, nil
}

type Source interface {
	SendTo(w io.WriteCloser)
}

func (p *Pipeline) CreateSink(name string, sink io.WriteCloser) (*DstVar, error) {
	v := &DstVar{
		Name:   name,
		Sink:   sink,
		waitCh: make(chan struct{}),
	}
	p.Sinks[name] = v
	return v, nil
}

func (p *Pipeline) CreateSpout(name string, src io.Reader) (*SrcVar, error) {
	conn := NewConnector()
	conn.ReadFrom(src)
	v := &SrcVar{
		Name:      name,
		Connector: conn,
		Spout:     src,
	}
	v.Token = p.Add(v)
	p.Spouts[name] = v
	return v, nil
}

func (p *Pipeline) ExitWhen(ctx context.Context, processOrVar string) error {
	proc, ok := p.Processes[processOrVar]
	if ok {
		_, err := proc.Wait(ctx, []ProcState{ProcFinished})
		p.Stop()
		return err
	}
	spout, ok := p.Sinks[processOrVar]
	if ok {
		spout.WaitClosed(ctx)
		p.Stop()
		return nil
	}
	return errMissingProcess(processOrVar)
}

func (p *Pipeline) Stop() {
	p.Creator.RemovePipeline(p)
}
