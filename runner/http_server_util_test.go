package runner_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Siroshun09/go-httplib"
	"github.com/Siroshun09/go-httplib/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	streamAPIChunkCount   = 5
	streamAPIChunkPrefix  = "chunk "
	streamAPILastChunkMsg = "last chunk"
)

func newTestHTTPServerRunner(t *testing.T) runner.HTTPServerRunner {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		httplib.RenderOK(r.Context(), w)
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		httplib.RenderOK(r.Context(), w)
	})

	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		fl, ok := w.(http.Flusher)
		if !ok {
			httplib.RenderInternalServerError(r.Context(), w, fmt.Errorf("not flusher"))
			return
		}

		for i := range streamAPIChunkCount {
			var msg string
			if i == streamAPIChunkCount-1 {
				msg = streamAPILastChunkMsg
			} else {
				msg = streamAPIChunkPrefix + strconv.Itoa(i)
			}

			_, err := w.Write([]byte(msg + "\n"))
			if err != nil {
				httplib.RenderInternalServerError(r.Context(), w, fmt.Errorf("failed to write chunk: %v", err))
				return
			}

			fl.Flush()
			time.Sleep(200 * time.Millisecond)
		}
	})

	return runner.NewHTTPServerRunner(
		&http.Server{
			Addr:    pickFreePort(t),
			Handler: mux,
		},
		func(ctx context.Context, err error) {
			require.FailNow(t, "server error occurred", "error: %+v", err)
		},
		func(ctx context.Context, rvr any) {
			require.FailNow(t, "panic occurred", "panic: %+v", rvr)
		},
	)
}

func pickFreePort(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "failed to listen for free port")
	defer func() {
		require.NoError(t, ln.Close(), "failed to close listener")
	}()
	addr := ln.Addr().String()
	return addr
}

var startServerLock sync.Mutex

func startTestHTTPServerRunner(ctx context.Context, t *testing.T) (s runner.HTTPServerRunner, baseURL string, srvCtx context.Context, stop func()) {
	t.Helper()

	startServerLock.Lock()
	defer startServerLock.Unlock()

	s = newTestHTTPServerRunner(t)

	srvCtx, stop = s.Run(ctx)
	baseURL = "http://" + s.Addr()
	waitHTTPServerReady(t, baseURL)

	return s, baseURL, srvCtx, stop
}

func waitHTTPServerReady(t *testing.T, baseURL string) {
	t.Helper()

	client := &http.Client{}
	start := time.Now()
	ticker := time.NewTicker(25 * time.Millisecond)
	timeout := 5 * time.Second

	defer ticker.Stop()
	for {
		resp, err := client.Get(baseURL + "/ok")
		if err == nil {
			require.NoError(t, resp.Body.Close())
			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		if timeout < time.Now().Sub(start) {
			require.Failf(t, "server did not become ready", "at %s", baseURL)
		}

		<-ticker.C
	}
}

func Test_TestHTTPServerRunner(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	assert.Equal(t, "http://"+s.Addr(), base)
	assert.NotNil(t, stop)

	assert.NoError(t, s.Shutdown(3*time.Second))
}
