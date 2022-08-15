package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSupervisorExits(t *testing.T) {
	s := New(t.TempDir())
	pipeline, err := s.AddPipeline("test")
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errch := s.ServeBackground(ctx)
	pipeline.Stop()

	assert.ErrorIs(t, <-errch, context.Canceled)
}
