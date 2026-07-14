# Scenario

**Feature**: Phase 4 per-session WebSocket isolation via SessionRegistry

```
# Registry-backed httptest (all leaves)
Test Client -> NewSessionRegistry(baseDir, addr)
Test Client -> NewRegistryControlHandler(registry) -> httptest.Server
Fake Extension -> GET /v1/ws?session=<id> hello|job|result
Test Client -> GET /v1/session?session= | POST /v1/jobs session_id=
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Phases 1–3 complete; Phase 4 hardens per-session WS routing (tests **RED** until
  implementer lands isolation).
- Tree root is `tests/browser-agent-daemon-phase4/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- All leaves use two pre-created sessions **SessionIDA** and **SessionIDB** in an
  isolated temp `BaseDir`; no real Chrome.
- WS client pattern copied from `tests/browser-agent/DOCTEST.md` with `?session=`
  on dial.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `SessionIDA = "sess-p4-a"`, `SessionIDB = "sess-p4-b"`.
4. Default hello version `1.0.0` and feature `browser-agent`.
5. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- After implementer, `doctest test ./tests/browser-agent/ws-control/...` must stay GREEN.

```go
import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	if req.SessionIDA == "" {
		req.SessionIDA = "sess-p4-a"
	}
	if req.SessionIDB == "" {
		req.SessionIDB = "sess-p4-b"
	}
	if req.HelloVersion == "" {
		req.HelloVersion = "1.0.0"
	}
	if req.HelloFeatures == nil {
		req.HelloFeatures = []string{"browser-agent"}
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertHTTPStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != want && resp.WSDialStatus != want {
		got := resp.StatusCode
		if resp.WSDialStatus != 0 {
			got = resp.WSDialStatus
		}
		t.Fatalf("HTTP status = %d, want %d; body=%s dialErr=%q",
			got, want, truncate(resp.BodyString, 400), resp.WSDialErr)
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

var _ = http.StatusOK
```