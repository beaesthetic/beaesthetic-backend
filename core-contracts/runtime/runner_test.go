package runtime

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunnerCancelsProcessesWhenCriticalProcessFails(t *testing.T) {
	ctx := context.Background()
	cancelled := make(chan struct{})

	runner := NewRunner(nil)
	runner.Add(Critical("failing", func(context.Context) error {
		return errors.New("boom")
	}))
	runner.Add(Critical("waiting", func(ctx context.Context) error {
		<-ctx.Done()
		close(cancelled)
		return ctx.Err()
	}))

	err := runner.Run(ctx)
	if err == nil {
		t.Fatal("Run() error is nil")
	}
	if !strings.Contains(err.Error(), "run failing: boom") {
		t.Fatalf("Run() error = %q", err)
	}
	select {
	case <-cancelled:
	default:
		t.Fatal("critical failure did not cancel other processes")
	}
}

func TestRunnerKeepsRunningWhenOptionalProcessFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	running := make(chan struct{})

	runner := NewRunner(nil)
	runner.Add(Optional("optional", func(context.Context) error {
		return errors.New("boom")
	}))
	runner.Add(Critical("critical", func(ctx context.Context) error {
		close(running)
		<-ctx.Done()
		return ctx.Err()
	}))

	complete := make(chan error, 1)
	go func() {
		complete <- runner.Run(ctx)
	}()

	<-running
	select {
	case err := <-complete:
		t.Fatalf("Run() completed before context cancellation: %v", err)
	default:
	}

	cancel()
	if err := <-complete; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
