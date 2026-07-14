## Expected

- First POST status **201** (recorded as `StatusCode`).
- Second POST status **409 Conflict**.
- Second body should indicate duplicate / already exists (soft: status mandatory).

## Side Effects

- Only one live session for `sess-dup`.

## Errors

- Second POST returning 201 or 400 fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first POST status=%d want 201", resp.StatusCode)
	}
	if resp.SecondPostStatus != http.StatusConflict {
		t.Fatalf("second POST status=%d want 409", resp.SecondPostStatus)
	}
}
```