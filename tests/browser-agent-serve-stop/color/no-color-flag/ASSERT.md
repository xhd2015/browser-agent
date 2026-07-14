## Expected

- `HandleCLI serve --no-color --stop` returns **nil** (exit **0**).
- Stderr contains **no** ANSI escape sequences (`\x1b`).

## Side Effects

- Read-only.

## Errors

- ANSI present in stderr or non-zero exit fails.

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
		t.Fatalf("HandleCLI exit = %d, want 0; cliErr=%q stderr=%q",
			resp.ExitCode, resp.CLIErr, truncate(resp.Stderr, 400))
	}
	if resp.HasANSI {
		t.Fatalf("expected no ANSI with --no-color; stderr=%q", truncate(resp.Stderr, 600))
	}
}
```