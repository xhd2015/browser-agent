## Expected

- `HandleCLI serve --color --stop` returns **nil** (exit **0**) against empty base-dir.
- Stderr contains at least one ANSI escape (`\x1b`).

## Side Effects

- Read-only (warning-only stop path).

## Errors

- Missing ANSI sequences or non-zero exit fails.

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
	if !resp.HasANSI {
		t.Fatalf("expected ANSI color in stderr with --color; stderr=%q",
			truncate(resp.Stderr, 600))
	}
}
```