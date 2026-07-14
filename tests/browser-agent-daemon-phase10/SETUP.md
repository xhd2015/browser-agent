# Scenario

**Feature**: Phase 10 docs & CLI help polish

```
fullHelp / briefUsage in cli.go document serve, serve --status, serve --kill-existing, session new
cmd/browser-agent/SKILL.md documents serve | session new → session info|eval|…
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Phase 10 help/skill docs not polished yet — tree is **RED**.
- Tree root is `tests/browser-agent-daemon-phase10/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Pure text contract tests only (no daemon spawn, no Chrome).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` and probe fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Help probes read `browseragent/cli.go` constants via `HandleCLI --help` or
  source extraction.
- Skill probe reads `cmd/browser-agent/SKILL.md` (or package mirror).

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
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
```