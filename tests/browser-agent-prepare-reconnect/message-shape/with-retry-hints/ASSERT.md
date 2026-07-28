## Expected

- Payload non-nil; `delay_ms` = 1000.
- `retry_base_ms` present and equals **50**.
- `retry_max_ms` present and equals **500**.
- `reason` = `daemon-upgrade`.

## Side Effects

- None.

## Errors

- Missing or wrong retry keys fails (extensions need hints for aggressive reconnect).

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
	if resp.DelayMSValue != 1000 {
		t.Fatalf("delay_ms=%d want 1000; payload=%s", resp.DelayMSValue, resp.PayloadJSON)
	}
	if !resp.HasRetryBase || resp.RetryBaseValue != 50 {
		t.Fatalf("retry_base_ms present=%v value=%d want 50; payload=%s",
			resp.HasRetryBase, resp.RetryBaseValue, resp.PayloadJSON)
	}
	if !resp.HasRetryMax || resp.RetryMaxValue != 500 {
		t.Fatalf("retry_max_ms present=%v value=%d want 500; payload=%s",
			resp.HasRetryMax, resp.RetryMaxValue, resp.PayloadJSON)
	}
	if resp.ReasonValue != "daemon-upgrade" {
		t.Fatalf("reason=%q want daemon-upgrade", resp.ReasonValue)
	}
	assertExitZero(t, resp)
}
```
