# Scenario

**Feature**: Phase 2 — prepare_reconnect WS envelope, broadcast, admin prepare-upgrade, extension markers

```
# pure payload shape
Test Client -> BuildPrepareReconnectPayload(opts)
  -> payload.delay_ms default 1000; reason / retry hints optional

# broadcast to live extension sockets
Fake Extension hello on /v1/ws?session=S
Test Client -> BroadcastPrepareReconnect(registry, opts)
  -> prepare_reconnect on WS; notified session ids

# admin HTTP
POST /v1/admin/prepare-upgrade -> broadcast -> { ok, notified }

# extension sources (Chrome + Firefox)
background.js handles prepare_reconnect -> delay_ms schedule close -> aggressive reconnect
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- Tree root is `tests/browser-agent-prepare-reconnect/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- **Classic TDD**: `BuildPrepareReconnectPayload`, `BroadcastPrepareReconnect`, and
  `POST /v1/admin/prepare-upgrade` are **not** GREEN yet; `doctest test` should be RED.
- Registry HTTP leaves use `t.TempDir()` BaseDir + `httptest` via
  `NewRegistryControlHandler`.
- Ext-source leaves are read-only filesystem probes (no browser).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` empty at root (grouping/leaf Setup sets Mode).
3. Shared helpers below are available to all leaves.

## Context

- Spec version **0.0.2**.
- Default delay: **1000** ms (`DefaultPrepareReconnectDelayMS`).
- Phase 3 (EnsureDaemon orchestration + 30s wait) is **out of scope**.

```go
import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	if req.HelloVersion == "" {
		req.HelloVersion = "1.0.0"
	}
	if req.HelloFeatures == nil {
		req.HelloFeatures = []string{"browser-agent"}
	}
	return nil
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

func sortedCopy(ids []string) []string {
	out := append([]string(nil), ids...)
	sort.Strings(out)
	return out
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as := sortedCopy(a)
	bs := sortedCopy(b)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

func assertNotifiedEquals(t *testing.T, got, want []string) {
	t.Helper()
	if !sameStringSet(got, want) {
		t.Fatalf("notified ids=%v want %v", got, want)
	}
}

func assertContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if !strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text to contain %q; got:\n%s", n, truncate(haystack, 900))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// hasPrepareReconnectTypeMarker reports type handling for prepare_reconnect.
func hasPrepareReconnectTypeMarker(text string) bool {
	if strings.Contains(text, "prepare_reconnect") {
		return true
	}
	// camelCase alternative still counts as intentional handling
	return strings.Contains(text, "prepareReconnect")
}

// hasDelayMsMarker reports reading delay_ms / delayMs from payload.
func hasDelayMsMarker(text string) bool {
	low := strings.ToLower(text)
	return strings.Contains(low, "delay_ms") || strings.Contains(low, "delayms")
}

// hasScheduleCloseMarker reports scheduling close after delay.
func hasScheduleCloseMarker(text string) bool {
	low := strings.ToLower(text)
	hasTimer := strings.Contains(low, "settimeout") || strings.Contains(low, "set_timeout") ||
		strings.Contains(low, "chrome.alarms") || strings.Contains(low, "alarms.create")
	hasClose := strings.Contains(low, ".close(") || strings.Contains(low, "ws.close") ||
		strings.Contains(low, "socket.close") || strings.Contains(low, "closews") ||
		strings.Contains(low, "close_ws") || strings.Contains(low, "close session")
	return hasTimer && hasClose
}

// hasAggressiveReconnectMarker reports post-close aggressive reconnect behavior.
func hasAggressiveReconnectMarker(text string) bool {
	low := strings.ToLower(text)
	if !strings.Contains(low, "reconnect") {
		return false
	}
	// Prefer explicit aggressive markers: reset attempt, force reconnect, lower base, prepare path.
	markers := []string{
		"reconnectattempt",
		"reconnect_attempt",
		"force reconnect",
		"forcereconnect",
		"aggressive",
		"retry_base",
		"retrybasems",
		"reconnect_base",
		"prepare_reconnect",
		"preparereconnect",
	}
	for _, m := range markers {
		if strings.Contains(low, m) {
			return true
		}
	}
	// Fallback: reconnect scheduled near close path (both tokens present).
	return strings.Contains(low, "connectsession") || strings.Contains(low, "schedule reconnect") ||
		strings.Contains(low, "schedulereconnect")
}
```
