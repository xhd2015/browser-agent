# Scenario

**Feature**: serve --stop + less-flags + cli-color

```
# stop path
RunDaemon -> server.json + health OK
HandleCLI serve --stop -> KillExistingDaemon -> health down -> meta gone -> exit 0

# idempotent stop
HandleCLI serve --stop (no meta) -> warning stderr -> exit 0

# parse / help / color
less-flags on cliServe; serve --help; cli-color on operator stderr
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Serve-stop feature (`serve --stop`, less-flags, cli-color, serve-specific help) not
  implemented yet — tree is **RED**.
- Tree root is `tests/browser-agent-serve-stop/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- All leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen when a
  daemon is needed.
- No real Chrome; no agent-run.
- **Env isolation**: `req.Env` is an explicit map passed to `HandleCLI` (never
  `nil` and never merged with process env). Ambient `NO_COLOR` from the host shell
  does not leak into color leaves. Only `color/no-color-env` re-injects
  `NO_COLOR=1`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Initialize `req.Env` to an empty map (explicit env only).
4. Default `ReadyTimeout = 5s`, `ShutdownWait = 8s`, `KillTimeout = 10s`.
5. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- Stop leaves use synchronous `HandleCLI` (`--stop` must not block like serve).
- Color leaves on pipe require `--color` to force ANSI (non-TTY).

```go
import (
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
	if req.Env == nil {
		req.Env = map[string]string{}
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.ShutdownWait == 0 {
		req.ShutdownWait = 8 * time.Second
	}
	if req.KillTimeout == 0 {
		req.KillTimeout = 10 * time.Second
	}
	return nil
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
```