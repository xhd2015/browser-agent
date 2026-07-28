## Expected

- Payload non-nil.
- `delay_ms` equals **2500** (not defaulted away).
- `reason` equals **`daemon-upgrade`**.

## Side Effects

- None.

## Errors

- Wrong delay or missing reason fails.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.Payload == nil {
		t.Fatal("payload is nil")
	}
	if resp.DelayMSValue != 2500 {
		t.Fatalf("delay_ms=%d want 2500; payload=%s", resp.DelayMSValue, resp.PayloadJSON)
	}
	if resp.ReasonValue != "daemon-upgrade" {
		t.Fatalf("reason=%q want %q; payload=%s", resp.ReasonValue, "daemon-upgrade", resp.PayloadJSON)
	}
	assertExitZero(t, resp)
}
```
