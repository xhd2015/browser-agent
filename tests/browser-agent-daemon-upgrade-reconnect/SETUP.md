# Scenario

**Feature**: Phase 3 — EnsureDaemon upgrade with prepare-upgrade, allow connected, wait ≤30s reattach

```
# pure waitList builder
Test Client -> ConnectedIDsFromSnapshots(views)
  -> [connected session ids]

# pure reattach wait (mock fetch)
Test Client -> WaitSessionsReconnected(waitList, fetch, timeout, poll)
  -> reattached[], stillWaiting[]

# EnsureDaemon orchestration (hooks)
client > daemon
  -> prepare-upgrade (best-effort) -> sleep delay+slack
  -> kill (keep dirs) -> spawn -> wait ready
  -> wait ≤30s reattach
  -> stderr upgraded + reattached / still waiting
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- Tree root is `tests/browser-agent-daemon-upgrade-reconnect/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- **Classic TDD**: Phase 3 helpers/hooks are **not** GREEN yet; `doctest test` should be RED.
- Pure leaves need no network. EnsureDaemon leaves use httptest health + injectable hooks.
- Default reattach wait product constant: **30s**; tests use short timeouts.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` empty at root (grouping/leaf Setup sets Mode).
3. Shared helpers below are available to all leaves.

## Context

- Spec version **0.0.2**.
- Depends on Phase 1 (keep session dirs) and Phase 2 (`POST /v1/admin/prepare-upgrade`).
- Sibling `ensure-daemon-upgrade/blocked-connected-warn-reuse` encodes pre-Phase-3 Q1 block.

```go
import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
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

func assertIDsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if !sameStringSet(got, want) {
		t.Fatalf("ids=%v want %v (order-insensitive set)", got, want)
	}
}

func assertIDsEqualOrdered(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ids=%v want ordered %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ids[%d]=%q want %q (full=%v)", i, got[i], want[i], got)
		}
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

func assertNotContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text NOT to contain %q; got:\n%s", n, truncate(haystack, 900))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func assertCallOrderHasSequence(t *testing.T, order []string, seq ...string) {
	t.Helper()
	if len(seq) == 0 {
		return
	}
	si := 0
	for _, step := range order {
		if step == seq[si] {
			si++
			if si == len(seq) {
				return
			}
		}
	}
	t.Fatalf("call order %v missing sequence %v", order, seq)
}

func assertHasSleepNear(t *testing.T, sleeps []time.Duration, want time.Duration, slack time.Duration) {
	t.Helper()
	for _, d := range sleeps {
		if d >= want-slack && d <= want+slack {
			return
		}
	}
	// Also accept single combined sleep of sum.
	t.Fatalf("SleepCalls %v missing duration near %s (±%s)", sleeps, want, slack)
}
```
