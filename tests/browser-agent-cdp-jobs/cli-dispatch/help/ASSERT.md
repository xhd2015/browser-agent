## Expected

Requirement **A1** (nested session CLI):

- HandleCLI error is **nil** (CLIErr empty).
- Printed help ends with `\n`.
- Lists top-level `serve` and `session`.
- Lists nested side-command names: `info`, `eval`, `run`, `logs`, `screenshot`, `cdp`
  (as substrings; may appear only under session).
- `DispatchTimedOut` false.

## Side Effects

- None.

## Errors

- Non-nil error or missing nested session commands is a failure.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("HandleCLI timed out on --help")
	}
	if resp.CLIErr != "" {
		t.Fatalf("--help should return nil error; got %q", resp.CLIErr)
	}
	assertExitZero(t, resp)
	assertPrintedTrailingNewline(t, resp)
	text := strings.ToLower(combinedCLIText(resp))
	if !strings.Contains(text, "serve") {
		t.Fatalf("help must list serve; got:\n%s", truncate(combinedCLIText(resp), 900))
	}
	// Nested complete refactor: require nested form, not incidental "session" in env names.
	if !strings.Contains(text, "session info") &&
		!strings.Contains(text, "session eval") &&
		!strings.Contains(text, "session run") {
		t.Fatalf("help must list nested session side-commands (e.g. session info/eval/run); got:\n%s",
			truncate(combinedCLIText(resp), 900))
	}
	for _, cmd := range []string{"info", "eval", "run", "logs", "screenshot", "cdp"} {
		if !strings.Contains(text, cmd) {
			t.Fatalf("help must list nested cmd %q; got:\n%s", cmd, truncate(combinedCLIText(resp), 900))
		}
	}
}
```
