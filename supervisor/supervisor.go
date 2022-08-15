package supervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/thejerf/suture/v4"
)

var (
	ErrAlreadyExists = errors.New("already exists with name")
)

type Supervisor struct {
	sup    *suture.Supervisor
	cancel func()

	Dir       string
	Pipelines map[string]*Pipeline
}

func New(dir string) *Supervisor {
	return &Supervisor{
		sup: suture.New("root", suture.Spec{
			EventHook: func(e suture.Event) {
				log.Debug().Msgf("%v", e)
			},
		}),
		Pipelines: make(map[string]*Pipeline),
		Dir:       dir,
	}
}

func (s *Supervisor) Serve(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)
	return s.sup.Serve(ctx)
}

func (s *Supervisor) ServeBackground(ctx context.Context) <-chan error {
	ctx, s.cancel = context.WithCancel(ctx)
	return s.sup.ServeBackground(ctx)
}

func (s *Supervisor) AddPipeline(name string) (*Pipeline, error) {
	if _, ok := s.Pipelines[name]; ok {
		return nil, fmt.Errorf("pipeline %s: %w", name, ErrAlreadyExists)
	}

	pipeline := NewPipeline(s, name, PipelineConfig{
		DataDir: filepath.Join(s.Dir, "pipelines", name),
	})
	pipeline.sid = s.sup.Add(pipeline)
	s.Pipelines[name] = pipeline
	return pipeline, nil
}

func (s *Supervisor) RemovePipeline(p *Pipeline) error {
	log.Debug().Str("pipeline", p.Name).Msg("stopping")
	err := s.sup.RemoveAndWait(p.sid, 10*time.Second)
	if err != nil {
		log.Warn().Str("pipeline", p.Name).Err(err).Msg("stopping pipeline failed")
		return err
	}

	delete(s.Pipelines, p.Name)
	if len(s.Pipelines) == 0 {
		s.cancel()
	}
	return nil
}

func (s *Supervisor) Close() error {
	s.cancel()
	return os.RemoveAll(s.Dir)
}
