## Expected

- `HandleCLI serve --stop --status` returns non-nil error (exit **1**).
- Error text mentions **`mutually exclusive`** or lists **`--stop`**, **`--status`**, and
  **`--kill-existing`**.

## Side Effects

- Read-only (no daemon started or stopped).

## Errors

- Nil CLI error or missing mutual-exclusion message fails.

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
		t.Fatal("expected non-nil CLI error for --stop --status")
	}
	if !resp.MutualExclSeen {
		t.Fatalf("error must mention mutually exclusive mode flags; cliErr=%q stderr=%q",
			resp.CLIErr, truncate(resp.Stderr, 600))
	}
}
```