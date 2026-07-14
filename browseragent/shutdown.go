package browseragent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const defaultKillExistingTimeout = 10 * time.Second

// runningDaemon tracks an in-process RunDaemon instance for force-stop when the
// recorded pid matches os.Getpid() (in-process doctest harness).
type runningDaemon struct {
	forceStop func()
}

var (
	runningDaemons   sync.Map // baseDir -> *runningDaemon
	defaultShutdownTO = 3 * time.Second
)

func registerRunningDaemon(baseDir string, d *runningDaemon) {
	runningDaemons.Store(filepath.Clean(baseDir), d)
}

func unregisterRunningDaemon(baseDir string) {
	runningDaemons.Delete(filepath.Clean(baseDir))
}

func lookupRunningDaemon(baseDir string) (*runningDaemon, bool) {
	v, ok := runningDaemons.Load(filepath.Clean(baseDir))
	if !ok {
		return nil, false
	}
	d, ok := v.(*runningDaemon)
	return d, ok
}

// ShutdownDaemon POSTs /v1/shutdown and polls GET /v1/health until down or timeout.
func ShutdownDaemon(baseURL string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultKillExistingTimeout
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return fmt.Errorf("shutdown: empty base URL")
	}

	status, body, err := postShutdown(baseURL)
	if err != nil {
		return fmt.Errorf("shutdown request: %w", err)
	}
	if status != http.StatusAccepted {
		return fmt.Errorf("shutdown: status %d: %s", status, strings.TrimSpace(string(body)))
	}

	if err := waitHealthDown(baseURL, timeout); err != nil {
		return err
	}
	return nil
}

// KillExistingDaemon reads server.json, requests graceful shutdown, waits up to
// killTimeout (default 10s), then force-kills the recorded pid if still alive.
// Stale server.json is removed when the daemon is gone.
func KillExistingDaemon(baseDir string, killTimeout time.Duration) error {
	if killTimeout <= 0 {
		killTimeout = defaultKillExistingTimeout
	}
	baseDir = filepath.Clean(baseDir)
	metaPath := filepath.Join(baseDir, "server.json")

	meta, err := ReadDaemonMeta(metaPath)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			return nil
		}
		return err
	}

	baseURL := strings.TrimSpace(meta.BaseURL)
	if baseURL == "" {
		baseURL = "http://" + strings.TrimSpace(meta.Addr)
	}
	if baseURL == "" || baseURL == "http://" {
		_ = RemoveDaemonMeta(metaPath)
		return nil
	}

	graceWait := killTimeout
	if graceWait > defaultKillExistingTimeout {
		graceWait = defaultKillExistingTimeout
	}

	// Best-effort graceful shutdown; continue even if POST fails (stale meta).
	_, _, _ = postShutdown(baseURL)

	deadline := time.Now().Add(graceWait)
	gracefulDown := false
	for time.Now().Before(deadline) {
		if daemonGracefullyDown(meta, baseDir, baseURL) {
			gracefulDown = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if gracefulDown {
		// Brief pause so concurrent observers (serve --kill-existing doctest)
		// can see connection refused before a replacement daemon re-binds the port.
		time.Sleep(50 * time.Millisecond)
	}

	needForce := isDaemonReachable(baseURL)
	if !needForce && meta.PID != os.Getpid() {
		needForce = IsProcessAlive(meta.PID)
	}

	if needForce {
		if err := forceKillDaemon(meta, baseDir, baseURL); err != nil {
			return err
		}
		waitRemain := killTimeout - time.Since(deadline)
		if waitRemain < 0 {
			waitRemain = 2 * time.Second
		}
		if waitRemain > 5*time.Second {
			waitRemain = 5 * time.Second
		}
		_ = waitHealthDown(baseURL, waitRemain)
	}

	return RemoveDaemonMeta(metaPath)
}

func daemonGracefullyDown(meta DaemonMeta, baseDir, baseURL string) bool {
	if isDaemonReachable(baseURL) {
		return false
	}
	if meta.PID != os.Getpid() {
		return !IsProcessAlive(meta.PID)
	}
	// In-process harness: wait until RunDaemon unregisters (clean exit).
	if _, ok := lookupRunningDaemon(baseDir); !ok {
		return true
	}
	return false
}

func forceKillDaemon(meta DaemonMeta, baseDir, baseURL string) error {
	if meta.PID == os.Getpid() {
		if inst, ok := lookupRunningDaemon(baseDir); ok && inst.forceStop != nil {
			inst.forceStop()
			return nil
		}
		// Fallback: try shutdown POST again.
		_, _, _ = postShutdown(baseURL)
		return nil
	}
	if err := KillProcess(meta.PID); err != nil {
		return fmt.Errorf("force kill pid %d: %w", meta.PID, err)
	}
	return nil
}

func postShutdown(baseURL string) (int, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/shutdown", nil)
	if err != nil {
		return 0, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, err
	}
	return res.StatusCode, body, nil
}

func isDaemonReachable(baseURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/health", nil)
	if err != nil {
		return false
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	return res.StatusCode == http.StatusOK
}

func waitHealthDown(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !isDaemonReachable(baseURL) {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("health still up at %s after %v", baseURL, timeout)
}

func shutdownResponseBody() []byte {
	data, _ := json.Marshal(map[string]bool{
		"ok":             true,
		"shutting_down":  true,
	})
	return append(data, '\n')
}

// shutdownAcceptedBody is the canonical POST /v1/shutdown JSON payload.
var shutdownAcceptedBody = shutdownResponseBody()

// shutdownDrain performs graceful HTTP server shutdown, honoring an optional
// pre-drain sleep when gracePeriod > 0 (force-kill doctest leaf).
func shutdownDrain(gracePeriod time.Duration, forceStop <-chan struct{}, httpSrv *controlHTTP) error {
	if gracePeriod > 0 {
		select {
		case <-time.After(gracePeriod):
		case <-forceStop:
			return httpSrv.srv.Close()
		}
	}
	shutdownTimeout := defaultShutdownTO
	if gracePeriod > 0 {
		shutdownTimeout = gracePeriod
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return httpSrv.shutdown(shutdownCtx)
}

