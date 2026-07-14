## Expected

- WS dial with `?session=sess-p4-a` succeeds.
- `WSDialOK` true.
- `WSDialStatus` is 101 (Switching Protocols) or 0 when handshake succeeded without response body.

## Side Effects

- Connection closed immediately after probe (harness does not send hello).

## Errors

- Dial error or non-upgrade status fails the leaf.

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
	if !resp.WSDialOK {
		t.Fatalf("WSDialOK=false; dialErr=%q status=%d", resp.WSDialErr, resp.WSDialStatus)
	}
}
```