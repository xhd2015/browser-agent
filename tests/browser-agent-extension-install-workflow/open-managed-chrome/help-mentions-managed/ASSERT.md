## Expected

- CLI help succeeds (nil error).
- Help text contains `open-managed-chrome`.
- Help text does not list `open-chrome` as a top-level command.

## Side Effects

- None.

## Errors

- Missing managed command in help fails.

## Exit Code

- 0.

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
		t.Fatalf("CLI error: %s", resp.CLIErr)
	}
	help := resp.Stdout
	assertContainsFold(t, help, "open-managed-chrome")
	assertNotContainsFold(t, help, "open-chrome [url]")
}
```