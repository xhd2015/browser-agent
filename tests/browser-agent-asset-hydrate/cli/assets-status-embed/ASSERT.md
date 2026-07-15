## Expected

- `HandleCLI` returns **nil** error.
- Stdout (case-insensitive) contains **session-page** and **extension**.
- Stdout mentions completeness somehow: at least one of `complete`, `incomplete`,
  `true`, `false`, `embed`, `cache` (status is about completeness).
- Stdout trailing `\n`.

## Side Effects

- No download required; cache dir may be referenced.

## Errors

- CLI error or missing kind tokens fails.

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
	if resp.CLIErr != nil {
		t.Fatalf("HandleCLI assets status err=%v stdout=%q stderr=%q",
			resp.CLIErr, truncate(resp.CLIStdout, 400), truncate(resp.CLIStderr, 200))
	}
	out := resp.CLIStdout
	low := strings.ToLower(out)
	if !strings.Contains(low, "session-page") {
		t.Fatalf("status missing session-page; out=%s", truncate(out, 500))
	}
	if !strings.Contains(low, "extension") {
		t.Fatalf("status missing extension; out=%s", truncate(out, 500))
	}
	// Completeness vocabulary
	hasCompleteVocab := strings.Contains(low, "complete") ||
		strings.Contains(low, "incomplete") ||
		strings.Contains(low, "true") ||
		strings.Contains(low, "false") ||
		strings.Contains(low, "embed") ||
		strings.Contains(low, "cache")
	if !hasCompleteVocab {
		t.Fatalf("status lacks completeness tokens (complete/embed/cache/…); out=%s",
			truncate(out, 500))
	}
	assertTrailingNewline(t, resp.CLIStdout, "CLIStdout")
	assertExitZero(t, resp)
}
```
