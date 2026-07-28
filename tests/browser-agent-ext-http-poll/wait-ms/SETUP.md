# Scenario

**Feature**: pure `NormalizeExtPollWaitMS` for poll long-poll bounds

```
# wait_ms normalization (no HTTP)
NormalizeExtPollWaitMS(input)
  -> default 25000 when <=0
  -> pass-through when 0 < input <= 30000
  -> cap 30000 when input > 30000
```

## Preconditions

- Nested root under `tests/browser-agent-ext-http-poll/wait-ms/`.
- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Helper **not** implemented yet → suite build RED until landed.

## Steps

1. Resolve ModuleRoot from DOCTEST_ROOT (two levels up from nested root → module).
2. Leaves set WaitMSCase + WaitMSInput.

## Context

- Constants: DefaultExtPollWaitMS=25000, MaxExtPollWaitMS=30000.
- Spec version **0.0.2**.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// Nested root: wait-ms → browser-agent-ext-http-poll → tests → module
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "..", ".."))
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
```
