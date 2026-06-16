package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type serverGroup struct {
	cancel    context.CancelFunc
	errs      chan error
	shutdowns []func(context.Context) error
	mu        sync.Mutex
}

func newServerGroup(parent context.Context) (*serverGroup, context.Context) {
	ctx, cancel := context.WithCancel(parent)
	return &serverGroup{
		cancel: cancel,
		errs:   make(chan error, 4),
	}, ctx
}

func (g *serverGroup) start(name string, serve func() error, shutdown func(context.Context) error) {
	g.mu.Lock()
	g.shutdowns = append(g.shutdowns, shutdown)
	g.mu.Unlock()

	go func() {
		if err := serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			g.errs <- fmt.Errorf("%s server: %w", name, err)
			g.cancel()
		}
	}()
}

func (g *serverGroup) addShutdown(shutdown func(context.Context) error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.shutdowns = append(g.shutdowns, shutdown)
}

func (g *serverGroup) wait(ctx context.Context, stderr io.Writer) error {
	select {
	case <-ctx.Done():
	case err := <-g.errs:
		g.shutdownAll(stderr)
		return err
	}

	g.shutdownAll(stderr)
	return ctx.Err()
}

func (g *serverGroup) shutdownAll(stderr io.Writer) {
	g.mu.Lock()
	shutdowns := append([]func(context.Context) error(nil), g.shutdowns...)
	g.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, shutdown := range shutdowns {
		if err := shutdown(ctx); err != nil {
			fmt.Fprintf(stderr, "shutdown error: %v\n", err)
		}
	}
}
