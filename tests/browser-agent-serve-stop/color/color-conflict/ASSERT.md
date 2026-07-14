## Expected

- `HandleCLI serve --color --no-color --stop` returns non-nil error (exit **1**).
- Error or stderr mentions the flags **cannot be specified together** or are otherwise
  mutually exclusive.

## Side Effects

- Read-only.

## Errors

- Nil CLI error or missing conflict message fails.

## Exit Code

- **1**.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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