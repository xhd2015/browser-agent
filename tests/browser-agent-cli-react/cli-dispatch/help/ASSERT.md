## Expected

Requirement **A2** (nested session CLI):

- HandleCLI error is **nil** (CLIErr empty).
- Printed help ends with `\n`.
- Lists top-level `serve` and `session`.
- Lists nested side-command names `info`, `eval` (as substrings; may appear only under session).
- `DispatchTimedOut` false.

## Side Effects

- None.

## Errors

- Non-nil error or os.Exit is a failure (harness would not observe output).
- Missing `session` means complete refactor incomplete.

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
		t.Fatalf("help must list serve; got:\n%s", truncate(combinedCLIText(resp), 800))
	}
	// Nested complete refactor: help must document session side-commands, not
	// merely the word "session" inside BROWSER_AGENT_SESSION_ID / prose.
	if !strings.Contains(text, "session info") && !strings.Contains(text, "session eval") {
		t.Fatalf("help must list nested session side-commands (e.g. session info/eval); got:\n%s",
			truncate(combinedCLIText(resp), 800))
	}
	for _, cmd := range []string{"info", "eval"} {
		if !strings.Contains(text, cmd) {
			t.Fatalf("help must list nested cmd %q; got:\n%s", cmd, truncate(combinedCLIText(resp), 800))
		}
	}
}
```
