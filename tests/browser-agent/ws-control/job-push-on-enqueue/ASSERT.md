## Expected

Requirement **D2**:

- `WSJobReceived` true.
- `WSJobType` is `eval`.

## Side Effects

- Background POST may still be waiting when leaf ends (server cancelled in cleanup).

## Errors

- No job envelope within harness timeout fails Run (transport error).

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
	if !resp.WSJobReceived {
		t.Fatal("WSJobReceived=false; expected type=job push after enqueue")
	}
	if resp.WSJobType != "eval" {
		t.Fatalf("WSJobType=%q, want eval; raw=%s", resp.WSJobType, truncate(resp.WSJobPayloadRaw, 300))
	}
}
```
