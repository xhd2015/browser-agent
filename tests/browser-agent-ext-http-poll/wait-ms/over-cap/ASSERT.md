## Expected

- `NormalizeExtPollWaitMS(99999)` returns **30000** (`MaxExtPollWaitMS`).

## Side Effects

- None.

## Errors

- Uncapped pass-through fails.

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
	const want = 30000
	if resp.NormalizedWaitMS != want {
		t.Fatalf("NormalizeExtPollWaitMS(99999)=%d want %d (MaxExtPollWaitMS)",
			resp.NormalizedWaitMS, want)
	}
	assertExitZero(t, resp)
}
```
