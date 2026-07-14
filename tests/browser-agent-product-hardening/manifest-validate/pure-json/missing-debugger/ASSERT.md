## Expected

Requirement **A2**:

- Validate fails (`ValidateOK` false).
- `ValidateErr` (case-insensitive) contains `debugger`.

## Side Effects

- None.

## Errors

- Silent OK on missing debugger is a failure.

## Exit Code

- Non-zero preferred (`ExitCode != 0`).

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
	if resp.ValidateOK {
		t.Fatal("expected validate failure when debugger permission missing")
	}
	errText := resp.ValidateErr
	if errText == "" {
		errText = resp.ErrText
	}
	if !strings.Contains(strings.ToLower(errText), "debugger") {
		t.Fatalf("validate error must mention debugger; got %q", errText)
	}
}
```
