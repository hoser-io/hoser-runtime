package supervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/hoser-io/hoser-runtime/hosercmd"
	"github.com/rs/zerolog/log"
	"github.com/thejerf/suture/v4"
)

type ProcState int

func (ps ProcState) String() string {
	switch ps {
	case ProcNotStarted:
		return "waiting"
	case ProcRunning:
		return "running"
	case ProcFinished:
		return "finished"
	case ProcError:
		return "error"
	default:
		return "invalid"
	}
}

const (
	ProcNotStarted ProcState = iota // Process is not started yet
	ProcFinished                    // Process is not running anymore (and won't be restarted)
	ProcRunning                     // Process is running still (actual OS process could be not alive temporarily)
	ProcError                       // Process is not running anymore because it had too many errors
)

// Processes live in a supervision tree that looks like:
//    Pipeline
//       |
//    ProcessSup
//    |         \
// Process    ValvesSup
//            /   |   \
//           V    V    V
// where V is a valve/connector and the process object actually contains
// the processup. Process is added to a supervisor via the Process.Supervise function.
// If Process crashes, ValvesSup is also restarted since all the named pipes need
// to be recreated.

type ProcessConfig struct {
	Argv       []string
	Ports      map[string]hosercmd.Port
	PrivateDir string
	SharedDir  string
}

func NewProcess(name, exePath string, cfg ProcessConfig) (*Process, error) {
	p := &Process{
		Name:    name,
		ExePath: exePath,
		Argv:    cfg.Argv,
		Ports:   cfg.Ports,
		DataDir: cfg.PrivateDir,

		stateNotify: make(chan struct{}, 1),

		Ins:  make(map[string]*InValve),
		Outs: make(map[string]*OutValve),
	}

	err := p.buildValves()
	if err != nil {
		return nil, fmt.Errorf("building valves: %w", err)
	}
	return p, nil
}

type ProcessSup struct {
	*suture.Supervisor
	procTok   suture.ServiceToken
	Valves    *suture.Supervisor // restarts valves if they fail for some reason.
	valvesTok suture.ServiceToken
}

func NewProcessSupervisor(proc *Process) *ProcessSup {
	ps := &ProcessSup{Supervisor: suture.New(proc.Name+"/sup", suture.Spec{
		DontPropagateTermination: true,
		EventHook: func(e suture.Event) {
			log.Debug().Str("supervisor", proc.Name).Msgf("%v", e)
		},
	})}
	ps.StartValves(proc)
	return ps
}

func (ps *ProcessSup) StartValves(proc *Process) {
	ps.Valves = suture.NewSimple(proc.Name + "/valves")
	for _, valve := range proc.Outs {
		ps.Valves.Add(valve)
	}
	ps.valvesTok = ps.Supervisor.Add(ps.Valves)
}

type ProcInfo struct {
	State ProcState
	Rc    int   // return code exited with
	Err   error // if exited with any error
}

func (pi ProcInfo) String() string {
	return fmt.Sprintf("{state: %v, rc: %d, err: %v}", pi.State, pi.Rc, pi.Err)
}

type Process struct {
	mu      sync.Mutex
	sup     *ProcessSup
	Token   suture.ServiceToken
	DataDir string // directory to store process specific data
	Name    string
	ExePath string
	Argv    []string
	Ports   map[string]hosercmd.Port

	Cmd         *exec.Cmd
	stateNotify chan struct{}
	Info        ProcInfo

	Ins  map[string]*InValve
	Outs map[string]*OutValve
}

// Supervise adds the process (a supervisor tree that manages the process) to
// sup.
func (p *Process) SupervisorTree() suture.Service {
	p.sup = NewProcessSupervisor(p)
	p.sup.Add(p)
	return p.sup
}

func (p *Process) buildValves() error {
	for name, port := range p.Ports {
		var err error
		if port.Dir == hosercmd.DirOut {
			_, err = p.AddOutValve(name)
		} else {
			_, err = p.AddInValve(name)
		}
		if err != nil {
			return err
		}
	}

	// Default valves for every process
	_, err := p.AddInValve(StdinValve)
	if err != nil {
		return err
	}

	_, err = p.AddOutValve(StdoutValve)
	if err != nil {
		return err
	}
	return nil
}

func (p *Process) buildCmd() (*exec.Cmd, error) {
	argv := make([]string, len(p.Argv)) // we need to do any variable substitution here
	var subs []string

	// $portname replaced with the filepath
	for name, port := range p.Ports {
		if port.Dir == hosercmd.DirOut {
			subs = append(subs, "$"+name, p.Outs[name].Path())
		} else {
			subs = append(subs, "$"+name, p.Ins[name].Path())
		}
	}

	r := strings.NewReplacer(subs...)
	for i := range p.Argv {
		argv[i] = r.Replace(p.Argv[i])
	}
	cmd := exec.Command(p.ExePath, argv...)
	cmd.Dir = p.DataDir

	// Need to OpenFile for stdio ports because we want the process we're starting to inherit the
	// open file descriptors instead of them being opened by name if we pass them in argv.
	var err error
	cmd.Stdin, err = p.Ins[StdinValve].OpenStdin()
	if err != nil {
		return nil, err
	}
	cmd.Stdout, err = p.Outs[StdoutValve].OpenStdout()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	return cmd, nil
}

func (p *Process) Serve(ctx context.Context) error {
	// err := os.MkdirAll(p.DataDir, 0755)
	// if err != nil {
	// 	return err
	// }
	// defer os.RemoveAll(p.DataDir)
	for _, valve := range p.Ins {
		valve.Open(ctx)
	}
	for _, valve := range p.Outs {
		valve.Open(ctx)
	}
	defer p.Close(ctx)

	var err error
	p.Cmd, err = p.buildCmd()
	if err != nil {
		return err
	}
	err = p.Cmd.Start()
	if err != nil {
		return err
	}
	p.ChangeState(func(pi *ProcInfo) { pi.State = ProcRunning })

	done := make(chan struct{})
	go p.monitorExit(ctx, done)
	defer close(done)

	err = p.Cmd.Wait()
	var sig syscall.Signal
	rc := 0
	if exerr, ok := err.(*exec.ExitError); ok {
		rc = exerr.ExitCode()
		if exerr.ProcessState != nil {
			status := exerr.ProcessState.Sys().(syscall.WaitStatus)
			sig = status.Signal()
		}
	}
	p.ChangeState(func(pi *ProcInfo) {
		pi.State = ProcFinished
		pi.Rc = rc
		pi.Err = err
	})
	if err == nil || sig == syscall.SIGHUP {
		// do not try to restart if clean exit of process (likely EOF)
		err = suture.ErrTerminateSupervisorTree
	}
	return err
}

func (p *Process) Close(ctx context.Context) error {
	for _, valve := range p.Ins {
		valve.Close()
	}
	for _, valve := range p.Outs {
		valve.Close()
	}
	return nil
}

// monitorExit waits for Process currently running to finish or context to end.
// If process is not exiting, try to kill with SIGHUP. If that does not succeed, print an error message and kill
// forcefully.
func (p *Process) monitorExit(ctx context.Context, finished chan struct{}) {
	select {
	case <-ctx.Done():
		log.Debug().Str("process", p.Name).Msg("sending SIGHUP")
		p.Cmd.Process.Signal(syscall.SIGHUP)
		// time.Sleep(5 * time.Second)
		// log.Debug().Str("process", p.Name).Msg("sending SIGKILL")
		// p.Cmd.Process.Kill()
	case <-finished:
	}
}

func (p *Process) IsFinished() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Info.State != ProcRunning
}

func (p *Process) ChangeState(modify func(*ProcInfo)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	oldState := p.Info.State
	modify(&p.Info)
	newState := p.Info.State
	log.Debug().Str("process", p.Name).Msgf("state: %v->%v", oldState, newState)
	if newState == ProcFinished {
		log.Debug().Str("process", p.Name).Int("rc", p.Info.Rc).Err(p.Info.Err).Msgf("finished")
	}
	if oldState != newState {
		select {
		case p.stateNotify <- struct{}{}:
		default: // do not block (drop any notifications if no one is listening)
		}
	}
}

func inState(wanted []ProcState, info ProcInfo) bool {
	for _, want := range wanted {
		if want == info.State {
			return true
		}
	}
	return false
}

func (p *Process) Wait(ctx context.Context, wanted []ProcState) (info ProcInfo, err error) {
	p.mu.Lock()
	info = p.Info
	if inState(wanted, info) {
		p.mu.Unlock()
		return // process is already in state
	}
	p.mu.Unlock()

	for !inState(wanted, info) {
		select {
		case <-p.stateNotify:
			p.mu.Lock()
			info = p.Info
			p.mu.Unlock()
		case <-ctx.Done():
			return ProcInfo{}, fmt.Errorf("waiting for process '%s' to enter states '%v', stuck in '%s': %w", p.Name, wanted, info.State, ctx.Err())
		}
	}
	return info, nil
}

func (p *Process) String() string {
	return p.Name
}
