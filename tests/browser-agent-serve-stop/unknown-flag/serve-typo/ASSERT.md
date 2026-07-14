## Expected

- `HandleCLI serve --foo` returns non-nil error (exit **1**).
- Stderr or error text mentions **`unrecognized flag`** or **`unknown flag`**
  (case-insensitive ok).

## Side Effects

- Read-only.

## Errors

- Nil CLI error or missing unknown-flag message fails.

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
		t.Fatal("expected non-nil CLI error for unknown serve flag")
	}
	if !resp.UnknownFlagSeen {
		t.Fatalf("error must mention unrecognized/unknown flag; cliErr=%q stderr=%q",
			resp.CLIErr, truncate(resp.Stderr, 600))
	}
}
```