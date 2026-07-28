# Scenario

**Feature**: Phase 1 — restore sessions from disk; upgrade keeps session dirs

```
# rehydrate registry from {baseDir}/sessions/*/meta.json
Test Client -> NewSessionRegistry -> RestoreSessionsFromDisk
  -> waiting_extension sessions in List/Get

# Create still exclusive (live or disk)
Test Client -> Create(id) -> ErrSessionExists when restored or disk-only

# upgrade must not wipe orphan session dirs
Test Client -> ensureDaemonKillAndRespawn(orphanIDs)
  -> SessionDirExists still true; preserved.txt intact

# RunDaemon wires restore before HTTP
seed meta -> RunDaemon -> GET /v1/sessions includes id
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- Tree root is `tests/browser-agent-daemon-session-restore/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- **Classic TDD**: `RestoreSessionsFromDisk` and upgrade no-wipe are **not** GREEN yet;
  `doctest test` should be RED.
- Disk tests use `t.TempDir()` via `ensureBaseDir`.
- Default **Addr** for pure registry leaves: `127.0.0.1:43761` (no listen required).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` empty at root (grouping/leaf Setup sets Mode).
3. Shared helpers below are available to all leaves.

## Context

- Spec version **0.0.2**.
- Phase constant: `browseragent.PhaseWaitingExtension` (`waiting_extension`).
- Sibling version-port leaf `upgrade-warn-orphan-dirs` currently expects wipe; product
  policy for Phase 1 is **keep dirs**.

```go
import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
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

func assertRestoreOK(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.RestoreErr != nil {
		t.Fatalf("RestoreSessionsFromDisk err=%v want nil", resp.RestoreErr)
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

func assertListContainsID(t *testing.T, resp *Response, id string) {
	t.Helper()
	for _, got := range resp.ListIDs {
		if got == id {
			return
		}
	}
	t.Fatalf("ListIDs %v missing %q", resp.ListIDs, id)
}

func assertListNotContainsID(t *testing.T, resp *Response, id string) {
	t.Helper()
	for _, got := range resp.ListIDs {
		if got == id {
			t.Fatalf("ListIDs %v unexpectedly contains %q", resp.ListIDs, id)
		}
	}
}

func assertWaitingExtension(t *testing.T, resp *Response, id string) {
	t.Helper()
	for _, e := range resp.ListEntries {
		if e.SessionID != id {
			continue
		}
		if e.Phase != browseragent.PhaseWaitingExtension {
			t.Fatalf("session %q phase=%q want %q", id, e.Phase, browseragent.PhaseWaitingExtension)
		}
		if e.ExtensionConnected {
			t.Fatalf("session %q extension.connected=true want false", id)
		}
		return
	}
	t.Fatalf("session %q not in ListEntries %v", id, resp.ListEntries)
}

func assertGetOK(t *testing.T, resp *Response, id string, want bool) {
	t.Helper()
	got, ok := resp.GetOK[id]
	if !ok {
		t.Fatalf("GetOK missing key %q", id)
	}
	if got != want {
		t.Fatalf("Get(%q)=%v want %v", id, got, want)
	}
}
```
