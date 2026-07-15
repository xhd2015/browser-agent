## Expected

- `HandleCLI` returns **nil** error.
- Combined stdout (and stderr if help routed there) contains **ensure** and
  **status** (case-insensitive OK).
- Stdout ends with trailing `\n`.

## Side Effects

- None (help only).

## Errors

- Non-nil CLI err, missing ensure/status tokens, or missing trailing newline fails.

## Exit Code

- 0 (nil HandleCLI error).

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
		t.Fatalf("HandleCLI assets --help err=%v stdout=%q stderr=%q",
			resp.CLIErr, truncate(resp.CLIStdout, 300), truncate(resp.CLIStderr, 200))
	}
	out := resp.CLIStdout + "\n" + resp.CLIStderr
	low := strings.ToLower(out)
	if !strings.Contains(low, "ensure") {
		t.Fatalf("assets help missing ensure; out=%s", truncate(out, 500))
	}
	if !strings.Contains(low, "status") {
		t.Fatalf("assets help missing status; out=%s", truncate(out, 500))
	}
	assertTrailingNewline(t, resp.CLIStdout, "CLIStdout")
	assertExitZero(t, resp)
}
```
