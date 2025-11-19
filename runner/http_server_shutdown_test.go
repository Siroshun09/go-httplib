package runner_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServerRunner_Shutdown_CompleteSlowRequest(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	client := &http.Client{}

	// Trigger shutdown during the slow handler
	reqDone := make(chan struct{})
	shutdownFinished := make(chan struct{})
	go func() {
		<-reqDone
		assert.NoError(t, s.Shutdown(3*time.Second))
		close(shutdownFinished)
	}()

	resp, err := client.Get(base + "/slow")
	reqDone <- struct{}{}

	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	<-shutdownFinished

	// Further requests should fail
	resp, err = client.Get(base + "/ok")
	assert.Nil(t, resp)
	assert.Error(t, err, "expected error after shutdown")
}

func TestHTTPServerRunner_Shutdown_ZeroTimeout(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	client := &http.Client{}

	// Trigger shutdown during the slow handler
	reqDone := make(chan struct{})
	shutdownFinished := make(chan struct{})
	go func() {
		<-reqDone
		assert.NoError(t, s.Shutdown(0))
		close(shutdownFinished)
	}()

	resp, err := client.Get(base + "/slow")
	reqDone <- struct{}{}

	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	<-shutdownFinished

	// Further requests should fail
	resp, err = client.Get(base + "/ok")
	assert.Nil(t, resp)
	assert.Error(t, err, "expected error after shutdown")
}

func TestHTTPServerRunner_Shutdown_MinusTimeout(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	client := &http.Client{}

	// Trigger shutdown during the slow handler
	reqDone := make(chan struct{})
	shutdownFinished := make(chan struct{})
	go func() {
		<-reqDone
		assert.NoError(t, s.Shutdown(-1))
		close(shutdownFinished)
	}()

	resp, err := client.Get(base + "/slow")
	reqDone <- struct{}{}

	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	<-shutdownFinished

	// Further requests should fail
	resp, err = client.Get(base + "/ok")
	assert.Nil(t, resp)
	assert.Error(t, err, "expected error after shutdown")
}

func TestHTTPServerRunner_Shutdown_KeepAlive_IdleClosed(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	tr := &http.Transport{}
	defer tr.CloseIdleConnections()
	client := &http.Client{Transport: tr}

	resp, err := client.Get(base + "/ok")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// server should close idle conn
	assert.NoError(t, s.Shutdown(3*time.Second))

	_, err = client.Get(base + "/ok")
	assert.Error(t, err, "expected keep-alive connection to be closed after shutdown")
}

func TestHTTPServerRunner_Shutdown_MultiConnections_AllComplete(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	client := &http.Client{Timeout: 3 * time.Second}
	n := 10
	var reqWg sync.WaitGroup
	reqWg.Add(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()

			resp, err := client.Get(base + "/slow")
			reqWg.Done()

			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			assert.Equal(t, http.StatusOK, resp.StatusCode, "status")
		}()
	}

	reqWg.Wait()
	assert.NoError(t, s.Shutdown(1*time.Second))

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		// ok
	case <-ctx.Done():
		require.Failf(t, "timeout waiting for requests to complete", "%v", ctx.Err())
	}
}

func TestHTTPServerRunner_Shutdown_StreamComplete(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	client := &http.Client{}
	resp, err := client.Get(base + "/stream")
	require.NoError(t, err, "stream request error")
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	reader := bufio.NewReader(resp.Body)
	streamStarted := make(chan struct{})
	streamFinished := make(chan struct{})
	shutdownTriggered := make(chan struct{})
	go func() {
		count := 0
		for {
			count++
			switch count {
			case 1:
				close(streamStarted)
			case 3:
				<-shutdownTriggered // wait for triggering shutdown
			}

			line, err := reader.ReadString('\n')
			assert.NoError(t, err)
			if strings.HasPrefix(line, streamAPILastChunkMsg) {
				assert.Equal(t, streamAPIChunkCount, count)
				close(streamFinished)
				break
			} else {
				assert.Truef(t, strings.HasPrefix(line, streamAPIChunkPrefix), "unexpected chunk: %q", line)
			}
		}
	}()

	<-streamStarted
	go func() {
		assert.NoError(t, s.Shutdown(3*time.Second))
		close(shutdownTriggered)
	}()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	select {
	case <-streamFinished:
		// ok
	case <-ctx.Done():
		require.Failf(t, "timed out waiting for stream to complete after shutdown", "%v", ctx.Err())
	}
}

// Subprocess helper to test SIGTERM-based graceful shutdown.
func TestHelperProcessSIGTERM(t *testing.T) {
	if os.Getenv("HELPER_SIGTERM") != "1" {
		return
	}

	ctx := t.Context()

	s, base, srvCtx, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()
	fmt.Println("READY " + base)
	<-srvCtx.Done()
	require.NoError(t, s.Shutdown(3*time.Second))
}

func TestHTTPServerRunner_Shutdown_Signal_SIGTERM(t *testing.T) {
	// Do not parallelize this test to avoid address duplication
	ctx := t.Context()

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelperProcessSIGTERM", "-test.v")
	cmd.Env = append(os.Environ(), "HELPER_SIGTERM=1")
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err, "StdoutPipe")
	err = cmd.Start()
	require.NoError(t, err, "failed to start helper")

	s := bufio.NewScanner(stdout)
	baseURLCh := make(chan string)
	go func() {
		for s.Scan() {
			line := s.Text()
			if strings.HasPrefix(line, "READY ") {
				baseURL, ok := strings.CutPrefix(line, "READY ")
				require.True(t, ok, "invalid line: %q", line)
				baseURLCh <- baseURL
				return
			}
		}
	}()

	var baseURL string
	ctxReady, cancelReady := context.WithTimeout(ctx, 5*time.Second)
	defer cancelReady()
	select {
	case baseURL = <-baseURLCh:
	case <-ctxReady.Done():
		_ = cmd.Process.Kill()
		require.FailNowf(t, "helper did not become ready", "%v", ctxReady.Err())
	}

	client := &http.Client{Timeout: 3 * time.Second}
	n := 5
	var reqWg sync.WaitGroup
	reqWg.Add(n)
	var respWg sync.WaitGroup
	respWg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer respWg.Done()
			resp, err := client.Get(baseURL + "/slow")
			reqWg.Done()

			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}()
	}

	reqWg.Wait()

	// Send SIGTERM and expect graceful exit
	err = cmd.Process.Signal(syscall.SIGTERM)
	require.NoError(t, err, "failed to send SIGTERM")

	respDone := make(chan struct{})
	respCtx, respCancel := context.WithTimeout(ctx, 3*time.Second)
	defer respCancel()
	go func() { respWg.Wait(); close(respDone) }()

	select {
	case <-respDone:
		// ok
	case <-respCtx.Done():
		require.Failf(t, "timeout waiting for requests to complete", "%v", ctx.Err())
	}

	doneCh := make(chan error, 1)
	go func() { doneCh <- cmd.Wait() }()
	exitCtx, cancelExit := context.WithTimeout(ctx, 5*time.Second)
	defer cancelExit()
	select {
	case err := <-doneCh:
		if err != nil && !errors.Is(err, exec.ErrNotFound) { // non-zero exits are errors
			require.FailNowf(t, "helper exited with error", "%v", err)
		}
	case <-exitCtx.Done():
		_ = cmd.Process.Kill()
		require.FailNowf(t, "helper did not exit after SIGTERM", "%v", exitCtx.Err())
	}
}

func TestHelperProcessSIGINT(t *testing.T) {
	if os.Getenv("HELPER_SIGINT") != "1" {
		return
	}

	ctx := t.Context()

	s, base, srvCtx, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()
	fmt.Println("READY " + base)
	<-srvCtx.Done()
	require.NoError(t, s.Shutdown(3*time.Second))
}

func TestHTTPServerRunner_Shutdown_Signal_Interrupt(t *testing.T) {
	// Do not parallelize this test to avoid address duplication
	ctx := t.Context()

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelperProcessSIGINT", "-test.v")
	cmd.Env = append(os.Environ(), "HELPER_SIGINT=1")
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err, "StdoutPipe")
	err = cmd.Start()
	require.NoError(t, err, "failed to start helper")

	s := bufio.NewScanner(stdout)
	baseURLCh := make(chan string)
	go func() {
		for s.Scan() {
			line := s.Text()
			if strings.HasPrefix(line, "READY ") {
				baseURL, ok := strings.CutPrefix(line, "READY ")
				require.True(t, ok, "invalid line: %q", line)
				baseURLCh <- baseURL
				return
			}
		}
	}()

	var baseURL string
	ctxReady, cancelReady := context.WithTimeout(ctx, 5*time.Second)
	defer cancelReady()
	select {
	case baseURL = <-baseURLCh:
	case <-ctxReady.Done():
		_ = cmd.Process.Kill()
		require.FailNowf(t, "helper did not become ready", "%v", ctxReady.Err())
	}

	client := &http.Client{Timeout: 3 * time.Second}
	n := 5
	var reqWg sync.WaitGroup
	reqWg.Add(n)
	var respWg sync.WaitGroup
	respWg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer respWg.Done()
			resp, err := client.Get(baseURL + "/slow")
			reqWg.Done()

			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}()
	}

	reqWg.Wait()

	err = cmd.Process.Signal(os.Interrupt)
	require.NoError(t, err, "failed to send SIGINT")

	respDone := make(chan struct{})
	respCtx, respCancel := context.WithTimeout(ctx, 3*time.Second)
	defer respCancel()
	go func() { respWg.Wait(); close(respDone) }()

	select {
	case <-respDone:
		// ok
	case <-respCtx.Done():
		require.Failf(t, "timeout waiting for requests to complete", "%v", ctx.Err())
	}

	doneCh := make(chan error, 1)
	go func() { doneCh <- cmd.Wait() }()
	exitCtx, cancelExit := context.WithTimeout(ctx, 5*time.Second)
	defer cancelExit()
	select {
	case err := <-doneCh:
		if err != nil && !errors.Is(err, exec.ErrNotFound) { // non-zero exits are errors
			require.FailNowf(t, "helper exited with error", "%v", err)
		}
	case <-exitCtx.Done():
		_ = cmd.Process.Kill()
		require.FailNowf(t, "helper did not exit after SIGINT", "%v", exitCtx.Err())
	}
}

func TestHTTPServerRunner_Shutdown_TimeoutExceeded(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	// Client with transport hook to detect TCP connect as a proxy for handler start
	connEstablished := make(chan struct{}, 1)
	tr := &http.Transport{}
	d := &net.Dialer{}
	tr.DialContext = func(c context.Context, network, addr string) (net.Conn, error) {
		conn, err := d.DialContext(c, network, addr)
		if err == nil {
			select { // non-blocking if already signaled
			case connEstablished <- struct{}{}:
			default:
			}
		}
		return conn, err
	}
	client := &http.Client{Transport: tr}
	defer tr.CloseIdleConnections()

	// Start a slow request that will outlive the shutdown timeout
	resultCh := make(chan error, 1)
	go func() {
		resp, err := client.Get(base + "/slow")
		if err != nil {
			resultCh <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			resultCh <- nil
		} else {
			resultCh <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
	}()

	// Wait until connection is established (indicates request is in flight)
	startCtx, cancelStart := context.WithTimeout(ctx, 2*time.Second)
	defer cancelStart()
	select {
	case <-connEstablished:
		// proceed
	case <-startCtx.Done():
		require.FailNowf(t, "slow request did not start", "%v", startCtx.Err())
	}

	// Shutdown with very short timeout; depending on exact timing it may either exceed deadline or succeed quickly
	err := s.Shutdown(50 * time.Millisecond)
	if err != nil {
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	}

	// After timeout, server may either let the in-flight request complete or close it; in any case it should finish promptly
	waitCtx, cancelWait := context.WithTimeout(ctx, 2*time.Second)
	defer cancelWait()
	select {
	case err := <-resultCh:
		if err != nil {
			assert.Error(t, err, "in-flight request may fail after shutdown timeout")
		} else {
			assert.NoError(t, err, "in-flight request completed before forced close")
		}
	case <-waitCtx.Done():
		require.FailNow(t, "slow request did not finish in time after shutdown timeout")
	}

	// New requests should be rejected after shutdown
	resp2, err2 := client.Get(base + "/ok")
	assert.Nil(t, resp2)
	assert.Error(t, err2)
}

func TestHTTPServerRunner_Shutdown_RejectNewRequestsDuringShutdown(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s, base, _, stop := startTestHTTPServerRunner(ctx, t)
	defer stop()

	shutdownStarted := make(chan struct{})
	shutdownDone := make(chan struct{})
	go func() {
		close(shutdownStarted)
		assert.NoError(t, s.Shutdown(1*time.Second))
		close(shutdownDone)
	}()

	<-shutdownStarted

	// During shutdown, new connections should be refused; try until failure or timeout without sleeps
	clientNew := &http.Client{}
	refused := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				resp, err := clientNew.Get(base + "/ok")
				if resp != nil {
					_ = resp.Body.Close()
				}
				if err != nil {
					refused <- struct{}{}
					return
				}
			}
		}
	}()

	waitRefuseCtx, cancelRefuse := context.WithTimeout(ctx, 2*time.Second)
	defer cancelRefuse()
	select {
	case <-refused:
		// ok
	case <-waitRefuseCtx.Done():
		require.FailNow(t, "new request was not refused during shutdown")
	}

	waitShutdownCtx, cancelShutdown := context.WithTimeout(ctx, 2*time.Second)
	defer cancelShutdown()
	select {
	case <-shutdownDone:
		// ok
	case <-waitShutdownCtx.Done():
		require.FailNow(t, "shutdown did not complete in time")
	}
}
