package runner

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HTTPServerRunner runs an http.Server with signal handling and graceful shutdown.
type HTTPServerRunner interface {
	// Addr returns the address set to the underlying http.Server.
	Addr() string
	// Run starts http.Server.ListenAndServe in a new goroutine.
	//
	// It returns a Context that will be canceled when syscall.SIGTERM or os.Interrupt is received,
	// and a stop function to stop signal notifications.
	//
	// NOTE: The input context's cancellation is intentionally removed via context.WithoutCancel.
	Run(ctx context.Context) (context.Context, func())
	// Shutdown gracefully shuts down the server.
	//
	// If timeout <= 0, it calls Server.Shutdown with context.Background(); otherwise
	// it uses context.WithTimeout(context.Background(), timeout).
	Shutdown(timeout time.Duration) error
}

// NewHTTPServerRunner creates an HTTPServerRunner for the given http.Server.
//
// Behavior:
// - Panics if server is nil.
// - If onError is nil, a no-op function is used.
// - If onPanic is nil, a no-op function is used.
func NewHTTPServerRunner(server *http.Server, onError func(ctx context.Context, err error), onPanic func(ctx context.Context, rvr any)) HTTPServerRunner {
	if server == nil {
		panic("server is nil")
	}

	if onError == nil {
		onError = func(ctx context.Context, err error) {}
	}

	if onPanic == nil {
		onPanic = func(ctx context.Context, rvr any) {}
	}

	return &httpServerRunner{
		server:  server,
		onError: onError,
		onPanic: onPanic,
	}
}

type httpServerRunner struct {
	server  *http.Server
	onError func(ctx context.Context, err error)
	onPanic func(ctx context.Context, rvr any)
}

func (r *httpServerRunner) Addr() string {
	return r.server.Addr
}

func (r *httpServerRunner) Run(ctx context.Context) (context.Context, func()) {
	ctx = context.WithoutCancel(ctx)
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, os.Interrupt)

	go func() {
		defer func() {
			if rvr := recover(); rvr != nil {
				r.onPanic(ctx, rvr)
			}
		}()

		if srvErr := r.server.ListenAndServe(); srvErr != nil {
			if !errors.Is(srvErr, http.ErrServerClosed) {
				r.onError(ctx, srvErr)
			}
		}
	}()

	return ctx, stop
}

func (r *httpServerRunner) Shutdown(timeout time.Duration) error {
	ctx := context.Background()

	if timeout <= 0 {
		return r.server.Shutdown(ctx)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return r.server.Shutdown(ctx)
}
