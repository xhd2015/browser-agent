# Scenario

**Feature**: Phase 1 daemon foundations — session id, daemon meta, session dir, process alive

```
# session id validation / generation
Test Client -> ValidateSessionID(id) / GenerateSessionID()
  -> nil or descriptive error / sess-xxxxxx

# daemon discovery file
Test Client -> WriteDaemonMeta(path, meta) -> ReadDaemonMeta(path)
  -> roundtrip pid, addr, base_url, base_dir, started_at

# session directory
Test Client -> SessionDirPath + SessionDirExists(baseDir, id)
  -> {baseDir}/sessions/{id}; bool from disk

# process liveness
Test Client -> IsProcessAlive(pid) -> true|false
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- Tree root is `tests/browser-agent-daemon-phase1/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- **Classic TDD**: APIs are not implemented yet; `doctest test` should be RED.
- Disk tests use `t.TempDir()` via `ensureBaseDir`; no HTTP or listen sockets.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` empty at root (grouping/leaf Setup sets Mode).
3. Shared helpers below are available to all leaves.

## Context

- Spec version **0.0.2**.
- Default base dir for docs only: `~/.tmp/browser-agent` (not used in tests).
- `DaemonMeta` JSON uses RFC3339 for `started_at`.

```go
import (
	"os"
	"path/filepath"
	"testing"
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
```