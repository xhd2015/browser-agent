## Expected

- `FormatDaemonStatus` returns **nil** error.
- Output contains **`Extension (embedded):`** with **`version`**, **`md5`**, **`path`**.
- Output contains **not running** status (not `Status:   running`).
- Output does **not** claim daemon is running.

## Side Effects

- Canonical extension extract under `TestHome` acceptable.

## Errors

- Missing extension block or running status marker fails.

## Exit Code

- Not asserted.

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
	if resp.FormatErr != "" {
		t.Fatalf("FormatDaemonStatus error: %s", resp.FormatErr)
	}
	if resp.Formatted == "" {
		t.Fatal("formatted output is empty")
	}
	if resp.Status.Running {
		t.Fatal("expected not-running status for format fixture")
	}
	assertContainsFold(t, resp.Formatted,
		"extension (embedded)",
		"version",
		"md5",
		"path",
		"not running",
	)
	low := strings.ToLower(resp.Formatted)
	if strings.Contains(low, "status:   running") || strings.Contains(low, "status:\trunning") {
		t.Fatalf("formatted output claims running; got:\n%s", truncate(resp.Formatted, 600))
	}
}
```