package runtime

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Runner struct {
	log       *zap.Logger
	processes []Process
}

func NewRunner(log *zap.Logger) *Runner {
	if log == nil {
		log = zap.NewNop()
	}
	return &Runner{log: log.Named("runtime_runner")}
}

func (r *Runner) Add(process Process) {
	r.processes = append(r.processes, process)
}

func (r *Runner) Run(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	for _, process := range r.processes {
		process := process
		group.Go(func() error {
			r.log.Info("starting process", zap.String("name", process.Name()))
			err := process.Run(groupCtx)
			if err == nil {
				if groupCtx.Err() != nil {
					r.log.Info("stopped process", zap.String("name", process.Name()))
					return nil
				}
				err = fmt.Errorf("process stopped unexpectedly")
			}
			if errors.Is(err, context.Canceled) && groupCtx.Err() != nil {
				r.log.Info("stopped process", zap.String("name", process.Name()))
				return nil
			}
			if !process.Critical() {
				r.log.Error("optional process stopped", zap.String("name", process.Name()), zap.Error(err))
				return nil
			}
			return fmt.Errorf("run %s: %w", process.Name(), err)
		})
	}
	return group.Wait()
}
