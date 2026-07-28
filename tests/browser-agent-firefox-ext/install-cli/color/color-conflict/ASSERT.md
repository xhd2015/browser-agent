## Expected

- `HandleCLI install-firefox-extension --color --no-color` returns non-nil CLI
  error (exit **1**).
- Error (or stderr) mentions the flags **cannot be specified together** (or are
  mutually exclusive) — same family of message as serve color conflict.

## Side Effects

- No requirement that extract runs (may fail at flag parse).

## Errors

- Nil CLI error, exit 0, or missing conflict message fails.

## Exit Code

- **1**.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	// Run returns nil transport error for this leaf; inspect resp.
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 1 {
		t.Fatalf("HandleCLI exit = %d, want 1; cliErr=%q stderr=%q",
			resp.ExitCode, resp.CLIErr, truncate(resp.Stderr, 400))
	}
	if resp.CLIErr == "" {
		t.Fatal("expected non-nil CLI error for --color --no-color")
	}
	if !resp.ColorConflictSeen {
		t.Fatalf("error must mention --color/--no-color conflict; cliErr=%q stderr=%q",
			resp.CLIErr, truncate(resp.Stderr, 600))
	}
}
```
