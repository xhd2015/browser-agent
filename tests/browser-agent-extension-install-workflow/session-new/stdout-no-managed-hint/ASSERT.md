## Expected

- `SessionNew` succeeds.
- Stdout does not mention `open-managed-chrome`.

## Side Effects

- None beyond stdout.

## Errors

- A managed-profile suggestion fails the test.

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
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	assertNotContainsFold(t, resp.Stdout, "open-managed-chrome")
}
```
