# Scenario

**Feature**: Phase 2 SessionRegistry — in-memory sessions with on-disk artifacts

```
# registry construction
Test Client -> NewSessionRegistry(baseDir, addr) -> *SessionRegistry

# create with artifacts
Test Client -> Create(id) -> CreateSessionResult | ErrSessionExists | validation error
  -> meta.json + SYSTEM.md under {baseDir}/sessions/{id}/

# lookup / enumerate / existence
Test Client -> Get(id) -> (*session, bool)
Test Client -> List() -> []sessionSnapshot sorted by id
Test Client -> Exists(id) -> registry OR disk dir
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- Phase 1 helpers (`ValidateSessionID`, `SessionDirPath`, `SessionDirExists`) exist.
- Tree root is `tests/browser-agent-daemon-phase2/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- **Classic TDD**: `SessionRegistry` is not implemented yet; `doctest test` should be RED.
- Disk tests use `t.TempDir()` via `ensureBaseDir`; no HTTP listen sockets.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Default **Addr** to `127.0.0.1:43761` when unset.
3. Shared helpers below are available to all leaves.

## Context

- Spec version **0.0.2**.
- `meta.json` omits `extension_install_path` in Phase 2 (deferred to Phase 5).
- `FormatSystemPrompt` does not embed the concrete session id in SYSTEM.md body.

```go
import (
	"errors"
	"net"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	return nil
}

func ensureBaseDir(t *testing.T, req *Request) string {
	t.Helper()
	if req.BaseDir == "" {
		req.BaseDir = t.TempDir()
	}
	return req.BaseDir
}

func ensureAddr(t *testing.T, req *Request) string {
	t.Helper()
	if req.Addr == "" {
		req.Addr = "127.0.0.1:43761"
	}
	return req.Addr
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d want 0", resp.ExitCode)
	}
}

func assertErrSessionExists(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("err=nil want ErrSessionExists")
	}
	if !errors.Is(err, browseragent.ErrSessionExists) {
		t.Fatalf("errors.Is(err, ErrSessionExists)=false got %v", err)
	}
}

func expectedBaseURL(addr string) string {
	return "http://" + addr
}

func expectedSessionURL(addr, sessionID string) string {
	return expectedBaseURL(addr) + "/go?session=" + sessionID
}

func expectedControlPort(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 43761
	}
	p, err := strconv.Atoi(portStr)
	if err != nil || p <= 0 {
		return 43761
	}
	return p
}
```