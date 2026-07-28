## Expected

- `NormalizeExtPollWaitMS(0)` returns **25000** (`DefaultExtPollWaitMS`).

## Side Effects

- None (pure helper).

## Errors

- Missing symbol / wrong default fails.

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
	const want = 25000
	if resp.NormalizedWaitMS != want {
		t.Fatalf("NormalizeExtPollWaitMS(0)=%d want %d (DefaultExtPollWaitMS)",
			resp.NormalizedWaitMS, want)
	}
	assertExitZero(t, resp)
}
```
