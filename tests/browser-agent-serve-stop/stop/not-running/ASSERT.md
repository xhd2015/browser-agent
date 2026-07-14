## Expected

- Empty `BaseDir` with no `server.json`.
- `HandleCLI serve --stop` returns **nil** (exit **0**).
- Stderr contains **`warning:`** and **`no daemon running`** (case-insensitive ok).

## Side Effects

- Read-only (no daemon started or stopped).

## Errors

- Non-nil CLI error or missing warning text fails.

## Exit Code

- **0**.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("HandleCLI exit = %d, want 0 for idempotent --stop; cliErr=%q stderr=%q",
			resp.ExitCode, resp.CLIErr, truncate(resp.Stderr, 400))
	}
	if !resp.WarningSeen {
		t.Fatalf("stderr missing idempotent warning; stderr=%q", truncate(resp.Stderr, 600))
	}
	assertContainsFold(t, resp.Stderr, "warning:", "no daemon running")
}
```