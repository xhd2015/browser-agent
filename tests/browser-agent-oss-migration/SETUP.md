# Scenario

**Feature**: browser-agent OSS migration repo contract (layout, module, hygiene, git)

```
# doctest harness inspects migrated browser-agent repo
Test Client -> DOCTEST_ROOT/../.. (repo root)
Test Client -> go build / go test / rg / git remote
Test Client <- pass/fail per migration leaf
```

## Preconditions

- Target repo root is `browser-agent/` (parent of `tests/`).
- **RepoRoot** = `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Classic TDD: repo is empty until implementer migrates; leaves are **RED** first.
- No real Chrome; shell-outs only to `go`, `rg`, `git`.

## Steps

1. Resolve `RepoRoot` from `DOCTEST_ROOT`.
2. Descendant Setup sets `Category` and `Leaf`.

## Context

- Parallel-safe: read-only inspection + isolated `go build`/`go test` in repo root.
- Shared helpers below are available to all descendant Assert packages.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport error: %v", err)
	}
}

func assertRunShellErr(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.RunErr == "" {
		return
	}
	t.Fatalf("command failed: %s\nstdout:\n%s\nstderr:\n%s",
		resp.RunErr, resp.Stdout, resp.Stderr)
}
```