# Scenario

**Feature**: script/github/release-assets packs three hydrate archives + help/docs mention --upload

```
# Pack-only (no GitHub) — explicit --out
Operator -> go run ./script/github/release-assets --out DIR --version v0.2.0
  -> DIR has AssetReleaseNames basenames (3 .tar.gz from embeds)
  -> exit 0

# Pack-only — default temp --out
Operator -> go run ./script/github/release-assets --version v0.2.0
  -> temp dir via MkdirTemp browser-agent-release-assets-*
  -> three archives; stdout: out: <abs-path>
  -> exit 0

# Help
Operator -> go run ./script/github/release-assets --help
  -> usage mentions --upload
  -> --out defaults to temp (not required)
  -> exit 0

# Operator docs (P7)
Test Client -> read docs/assets-hydrate.md
  -> mentions script/github/release-assets + --upload
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root `tests/browser-agent-release-assets/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Embed sources must exist on disk under ModuleRoot (already committed):
  - `browseragent/embedded/session-page`
  - `browseragent/embedded/extension`
  - `browsertrace/embedded/extension`
- Script path under test: `script/github/release-assets` (package main).
- Preferred operator doc: `docs/assets-hydrate.md`.
- Leaves never pass `--upload` and never require network or `gh`.
- `browseragent.AssetReleaseNames` already exists (P7 name helper).
- Avoid leaf slugs ending in `-js`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` empty at root (grouping/leaf sets it).
3. Shared helpers for all leaves.

## Context

- Spec version **0.0.3**.
- Parallel-safe: per-leaf temp `--out` dirs (explicit) or script-owned temp dirs
  (`PackOmitOut`); docs leaves are read-only FS.
- `go run` may compile; 3-minute timeout in `Run` for pack/help.
- Classic TDD for default-temp-out / help out-default-temp until implementer GREEN.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	if st, err := os.Stat(req.ModuleRoot); err != nil || !st.IsDir() {
		t.Fatalf("ModuleRoot %s not a directory: %v", req.ModuleRoot, err)
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d want 0; err=%q stdout=%s stderr=%s",
			resp.ExitCode, resp.ErrText,
			truncate(resp.Stdout, 400), truncate(resp.Stderr, 400))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func assertTrailingNewline(t *testing.T, s, label string) {
	t.Helper()
	if s == "" || !strings.HasSuffix(s, "\n") {
		t.Fatalf("%s must end with trailing newline; got %q", label, truncate(s, 200))
	}
}

func combinedOut(resp *Response) string {
	if resp == nil {
		return ""
	}
	return resp.Stdout + resp.Stderr
}
```
