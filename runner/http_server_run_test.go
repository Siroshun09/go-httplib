package runner_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Siroshun09/go-httplib/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServerRunner_Run_onError_called_when_listen_fails(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, _, _, stop1 := startTestHTTPServerRunner(ctx, t)
	defer stop1()
	defer func() {
		require.NoError(t, s.Shutdown(1*time.Second))
	}()

	errCh := make(chan error, 1)

	r := runner.NewHTTPServerRunner(
		&http.Server{Addr: s.Addr(), Handler: http.NewServeMux()},
		func(ctx context.Context, err error) { errCh <- err },
		func(ctx context.Context, rvr any) { require.FailNow(t, "unexpected panic", "%+v", rvr) },
	)

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, stop2 := r.Run(timeoutCtx)
	defer stop2()

	select {
	case err := <-errCh:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bind: address already in use")
	case <-timeoutCtx.Done():
		require.FailNow(t, "onError was not called")
	}
}

func TestHTTPServerRunner_Run_nil_onError_ignored_when_listen_fails(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, _, _, stop1 := startTestHTTPServerRunner(ctx, t)
	defer stop1()
	defer func() {
		require.NoError(t, s.Shutdown(1*time.Second))
	}()

	r := runner.NewHTTPServerRunner(
		&http.Server{Addr: s.Addr(), Handler: http.NewServeMux()},
		nil, // onError is nil; should be safely ignored
		func(ctx context.Context, rvr any) { require.FailNow(t, "unexpected panic", "%+v", rvr) },
	)

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, stop2 := r.Run(timeoutCtx)
	defer stop2()

	<-timeoutCtx.Done()
}

func TestHTTPServerRunner_Run_onPanic_called_when_panic_occurs(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, _, _, stop1 := startTestHTTPServerRunner(ctx, t)
	defer stop1()
	defer func() {
		require.NoError(t, s.Shutdown(1*time.Second))
	}()

	panicCh := make(chan any, 1)

	r := runner.NewHTTPServerRunner(
		&http.Server{Addr: s.Addr(), Handler: http.NewServeMux()},
		func(ctx context.Context, err error) { panic(err) },
		func(ctx context.Context, rvr any) { panicCh <- rvr },
	)

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, stop2 := r.Run(timeoutCtx)
	defer stop2()

	select {
	case p := <-panicCh:
		assert.NotNil(t, p)

		err, ok := p.(error)
		assert.True(t, ok)
		assert.Contains(t, err.Error(), "bind: address already in use")
	case <-timeoutCtx.Done():
		require.FailNow(t, "onPanic was not called")
	}
}

func TestHTTPServerRunner_Run_nil_onPanic_ignored_when_panic_occurs(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, _, _, stop1 := startTestHTTPServerRunner(ctx, t)
	defer stop1()
	defer func() {
		require.NoError(t, s.Shutdown(1*time.Second))
	}()

	r := runner.NewHTTPServerRunner(
		&http.Server{Addr: s.Addr(), Handler: http.NewServeMux()},
		func(ctx context.Context, err error) { panic(err) },
		nil, // onPanic is nil; panic should be recovered and ignored
	)

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, stop2 := r.Run(timeoutCtx)
	defer stop2()

	<-timeoutCtx.Done()
}
