## Expected Output

```text
Version:
Extension (embedded):
md5
path
```

## Expected

- `FormatDaemonStatus` returns **nil** error.
- Output contains **`Version:`** after uptime section.
- Output contains **`Extension (embedded):`** block with **`version`**, **`md5`**, **`path`** lines.
- Output still contains running status markers (`Status:`, `Sessions`).

## Side Effects

- Formatter writes only to provided writer.

## Errors

- Format error or missing rich markers fails.

## Exit Code

- Not asserted.

```go
import (
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
	if !resp.Status.Running {
		t.Fatal("expected running status for format fixture")
	}
	assertContainsFold(t, resp.Formatted,
		"version:",
		"extension (embedded)",
		"md5",
		"path",
		"status:",
		"sessions",
	)
}
```