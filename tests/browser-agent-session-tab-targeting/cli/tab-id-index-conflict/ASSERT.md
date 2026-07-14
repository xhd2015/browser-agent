## Expected

- `HandleCLI` returns non-nil error (exit **1**).
- Error or stderr mentions **cannot use both --tab-id and --tab-index** (or equivalent
  mutual-exclusion wording referencing both flags).
- Job must **not** be posted (conflict leaf has no server; parse-time failure).
- Until implementer lands flags, ignored flags may surface a different error — still **RED**.

## Side Effects

- Read-only CLI dispatch.

## Errors

- Nil CLI error, success exit, or missing conflict message fails.

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
		t.Fatalf("ExitCode=%d want 1 for --tab-id + --tab-index; cliErr=%q stderr=%q stdout=%q",
			resp.ExitCode, resp.CLIErr, truncate(resp.Stderr, 400), truncate(resp.Stdout, 200))
	}
	if resp.CLIErr == "" {
		t.Fatal("expected non-nil CLI error for conflicting tab flags")
	}
	if !resp.TabIDConflictSeen {
		t.Fatalf("error must mention --tab-id/--tab-index conflict; cliErr=%q stderr=%q",
			resp.CLIErr, truncate(resp.Stderr, 600))
	}
}
```