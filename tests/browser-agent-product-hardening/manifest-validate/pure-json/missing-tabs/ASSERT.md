## Expected

Requirement **A3**:

- Validate fails (`ValidateOK` false).
- `ValidateErr` (case-insensitive) contains `tabs`.

## Side Effects

- None.

## Errors

- Silent OK on missing tabs is a failure.

## Exit Code

- Non-zero preferred.

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
		t.Fatal("expected validate failure when tabs permission missing")
	}
	errText := resp.ValidateErr
	if errText == "" {
		errText = resp.ErrText
	}
	if !strings.Contains(strings.ToLower(errText), "tabs") {
		t.Fatalf("validate error must mention tabs; got %q", errText)
	}
}
```
