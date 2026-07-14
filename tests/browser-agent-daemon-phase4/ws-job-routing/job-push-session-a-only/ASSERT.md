## Expected

- `WSJobReceivedOnA` true.
- `WSJobType` is `eval`.
- `WSJobReceivedOnB` false (B socket must not receive A's job).

## Side Effects

- Background POST on A may still be waiting when leaf ends (cleanup closes server).

## Errors

- No job on A within timeout fails Run (transport error).
- Job received on B fails isolation.

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
	if !resp.WSJobReceivedOnA {
		t.Fatal("WSJobReceivedOnA=false; expected type=job on session A socket")
	}
	if resp.WSJobType != "eval" {
		t.Fatalf("WSJobType=%q want eval; raw=%s", resp.WSJobType, truncate(resp.WSJobPayloadRaw, 300))
	}
	if resp.WSJobReceivedOnB {
		t.Fatal("WSJobReceivedOnB=true; job for A must not be delivered on B socket")
	}
}
```