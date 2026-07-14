# Scenario

**Feature**: Session tab targeting (--tab-id / --tab-index) across CLI, extension, info, e2e

```
Operator -> HandleCLI session eval|… --tab-id|--tab-index
  -> POST /v1/jobs { tab_id, type, params }
Extension -> validate tab in session window; attach debugger; eval/screenshot
Operator -> session info (human table + --json job_target)
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- **ModuleRoot** = `filepath.Join(DOCTEST_ROOT, "..", "..")`.
- Daemon phases 1–9 (`RunDaemon`, `POST /v1/sessions`, per-session extension WS).
- Classic TDD: feature **not** implemented — static/cli/info leaves **RED** expected.
- E2e leaves: `playwright-debug` on PATH; skip when absent; **RED** when present.
- Reuses harness patterns from `browser-agent-e2e-playwright` and
  `browser-agent-active-tab-routing`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Grouping Setup sets `Mode` (`cli`, `ext-source`, `info`, `e2e`).
3. Server-backed modes allocate temp `BaseDir` and default timeouts.
4. Leaf Setup sets `CLIOp`, `ExtSourceTarget`, `InfoOp`, or `PlaywrightOp`.

## Context

- Spec version **0.0.2**.
- No `openChrome` — sessions created only via `POST /v1/sessions`.
- Fake extension dials `/v1/ws?session=<id>` for CLI/info leaves.
- E2e ASSERT frontmatter: `slow, ui-automation`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 10 * time.Second
	}
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 12 * time.Second
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func hasTabIDWindowValidation(text string) bool {
	low := strings.ToLower(text)
	// Explicit job tab_id path — not satisfied by pickTargetTabIdForSession alone.
	hasJobTabID := (strings.Contains(low, "tab_id") || strings.Contains(low, "tabid")) &&
		(strings.Contains(low, "job.tab_id") || strings.Contains(low, "payload.tab_id") ||
			strings.Contains(low, "params.tab_id") || strings.Contains(low, "resolvetargettab") ||
			strings.Contains(low, "resolve_target_tab") || strings.Contains(low, "picktargettabidforsession(sessionid, tabid"))
	hasWindowCheck := strings.Contains(low, "entry.windowid") &&
		(strings.Contains(low, "tabs.get") || strings.Contains(low, "windowid:") || strings.Contains(low, "windowid :"))
	return hasJobTabID && hasWindowCheck
}

func hasTabIndexOrderResolution(text string) bool {
	low := strings.ToLower(text)
	hasIndex := strings.Contains(low, "tab_index") || strings.Contains(low, "tabindex") ||
		strings.Contains(low, "tab-index")
	hasCapturable := strings.Contains(low, "capturable") || strings.Contains(low, "iscapturable")
	hasWindowTabs := strings.Contains(low, "windowid") && strings.Contains(low, "tabs.query")
	// 1-based: index - 1 or index === n patterns
	hasOneBased := strings.Contains(low, "index - 1") || strings.Contains(low, "index-1") ||
		strings.Contains(low, "1-based") || strings.Contains(low, "tab_index_flag")
	return hasIndex && (hasCapturable || hasWindowTabs) && hasOneBased
}

func hasAttachReuseAndDetach(text string) bool {
	low := strings.ToLower(text)
	hasReuse := strings.Contains(low, "attachedtabs.has") || strings.Contains(low, "already attached")
	hasDetach := strings.Contains(low, "debugger.detach") || strings.Contains(low, "chrome.debugger.detach")
	hasSwitch := strings.Contains(low, "tab_id") && (strings.Contains(low, "switch") || strings.Contains(low, "different"))
	hasSerialize := strings.Contains(low, "attachlock") || strings.Contains(low, "attachmutex") ||
		strings.Contains(low, "attachqueue") || strings.Contains(low, "serializ")
	return hasReuse && hasDetach && (hasSwitch || hasSerialize)
}
```