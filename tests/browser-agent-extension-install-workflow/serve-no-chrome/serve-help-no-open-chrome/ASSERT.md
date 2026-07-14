## Expected

- `serve --help` succeeds.
- Help text does **not** contain `--no-open-chrome`.

## Side Effects

- None.

## Errors

- Presence of obsolete flag in help fails.

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
	assertNotContainsFold(t, resp.Stdout, "--no-open-chrome")
}
```