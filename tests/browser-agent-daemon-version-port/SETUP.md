# Scenario

**Feature**: daemon version, fixed port, host/port flags, upgrade semantics

```
Client CLI -> EnsureDaemon / serve / session new
Daemon Host -> GET /v1/health + server.json (version + base_dir)
Fake Extension -> WS hello (connected gate for upgrade)
Foreign Listener -> non-browser-agent HTTP on control port
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Version-port feature largely absent — tree is **Classic TDD RED** until implementer
  lands `VERSION.txt`, `version.go`, inject hooks, extended health, flag migration,
  upgrade logic, foreign-port detection, TTY `serve --stop`.
- Tree root is `tests/browser-agent-daemon-version-port/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Leaves use isolated temp `BaseDir`; binds use `pickFreePort(t)` → `127.0.0.1:N`
  (avoid parallel collisions on product port **43761**).
- Default constants (`DefaultAddr`, `DefaultControlPort`) asserted in unit/help leaves,
  not by binding 43761 in every integration leaf.
- No real Chrome; no agent-run; fake extension WS pattern from phase4.
- **Version injection** (implementer): `browseragent/inject/versionhooks.go` —
  `ClientVersionOverride`, `DaemonVersionForHealth`; tests set/clear via `defer`.
- **TTY injection** (implementer): `browseragent/inject/terminalhooks.go` —
  `IsTerminalFn(io.Reader) bool` for `serve --stop` confirm leaves.
- **Sibling tree updates** (implementer, not this tree): phase8 spawn paths need explicit
  `--port`; phase10 help; `browser-agent-session-addr-resolve` for `--server-port`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `ShutdownWait = 8s`, `SessionIDA = "sess-vp-a"`,
   `SessionIDB = "sess-vp-b"`, `ClientVersion = "0.2.0"`, `DaemonVersion = "0.1.0"`.
4. Initialize `req.Env` to empty map when unset (explicit env for CLI leaves).
5. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- Loose semver compare ignores pre-release suffix for ordering; missing daemon version = `0.0.0`.
- Upgrade gate counts only `extension.connected == true` (not `waiting_extension` alone).

```go
import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if req.Env == nil {
		req.Env = map[string]string{}
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.ShutdownWait == 0 {
		req.ShutdownWait = 8 * time.Second
	}
	if req.SessionIDA == "" {
		req.SessionIDA = "sess-vp-a"
	}
	if req.SessionIDB == "" {
		req.SessionIDB = "sess-vp-b"
	}
	if req.ClientVersion == "" {
		req.ClientVersion = "0.2.0"
	}
	if req.DaemonVersion == "" {
		req.DaemonVersion = "0.1.0"
	}
	if req.HelloVersion == "" {
		req.HelloVersion = "1.0.0"
	}
	if req.HelloFeatures == nil {
		req.HelloFeatures = []string{"browser-agent"}
	}
	return nil
}

func pickFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("pickFreePort: %v", err)
	}
	_, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		_ = ln.Close()
		t.Fatalf("pickFreePort split: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		_ = ln.Close()
		t.Fatalf("pickFreePort atoi: %v", err)
	}
	_ = ln.Close()
	return port
}

func loopbackAddr(port int) string {
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if !strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func assertNotContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text NOT to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func daemonMetaPath(baseDir string) string {
	return filepath.Join(baseDir, "server.json")
}

func sessionDirPath(baseDir, sessionID string) string {
	return filepath.Join(baseDir, "sessions", sessionID)
}
```