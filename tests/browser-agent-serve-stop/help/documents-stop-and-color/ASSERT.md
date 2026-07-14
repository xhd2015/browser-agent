## Expected

- `HandleCLI serve --help` returns **nil** (`CLIErr` empty).
- Help text (stdout) contains **`--stop`**, **`--color`**, and **`--no-color`**.
- Does **not** require the full top-level session command tree in this leaf.

## Side Effects

- Read-only.

## Errors

- Missing any of the three flags in help text fails.

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
	if resp.CLIErr != "" {
		t.Fatalf("serve --help should return nil error; got %q", resp.CLIErr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("serve --help exit = %d, want 0", resp.ExitCode)
	}
	assertContainsFold(t, resp.HelpText, "--stop", "--color", "--no-color")
}
```