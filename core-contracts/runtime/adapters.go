package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type ContextRunner interface {
	Run(context.Context) error
}

func Consumer(name string, consumer ContextRunner) Process {
	return Critical(name, consumer.Run)
}

func HTTPServer(name string, server *http.Server, shutdownTimeout time.Duration) Process {
	return Critical(name, func(ctx context.Context) error {
		serverErr := make(chan error, 1)
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				serverErr <- err
				return
			}
			serverErr <- nil
		}()

		select {
		case err := <-serverErr:
			return err
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("shutdown http server: %w", err)
			}
			return <-serverErr
		}
	})
}

func StartStop(name string, start func(context.Context) error, stop func(context.Context) error, shutdownTimeout time.Duration) Process {
	return Critical(name, func(ctx context.Context) error {
		if err := start(ctx); err != nil {
			return err
		}
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return stop(shutdownCtx)
	})
}
