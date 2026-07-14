## Expected

- WS dial without `session` fails with HTTP **400** (not 404, not upgrade).
- `WSDialOK` false.
- `WSDialStatus` = 400.

## Side Effects

- No WebSocket connection established.

## Errors

- 404 or successful upgrade fails the leaf.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.WSDialOK {
		t.Fatal("WSDialOK=true; expected dial failure without session param")
	}
	if resp.WSDialStatus != http.StatusBadRequest {
		t.Fatalf("WSDialStatus=%d want %d; dialErr=%q",
			resp.WSDialStatus, http.StatusBadRequest, resp.WSDialErr)
	}
}
```