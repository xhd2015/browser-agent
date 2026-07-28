## Expected

- `BuildPrepareReconnectPayload` returns a non-nil map.
- `payload["delay_ms"]` is **1000** (`DefaultPrepareReconnectDelayMS`).
- `reason` is absent or empty when not provided.
- Retry hint keys are **not** required when unset.

## Side Effects

- None (pure function).

## Errors

- Nil payload or wrong default delay fails this leaf.

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
		t.Fatal("payload is nil; BuildPrepareReconnectPayload must return a map")
	}
	if resp.DelayMSValue != 1000 {
		t.Fatalf("delay_ms=%d want 1000 (default); payload=%s", resp.DelayMSValue, resp.PayloadJSON)
	}
	if resp.ReasonValue != "" {
		t.Fatalf("reason=%q want empty when not provided", resp.ReasonValue)
	}
	assertExitZero(t, resp)
}
```
