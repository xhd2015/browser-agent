## Expected

- `NormalizeExtPollWaitMS(5000)` returns **5000**.

## Side Effects

- None.

## Errors

- Accidental defaulting or capping fails.

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
	if resp.NormalizedWaitMS != 5000 {
		t.Fatalf("NormalizeExtPollWaitMS(5000)=%d want 5000", resp.NormalizedWaitMS)
	}
	assertExitZero(t, resp)
}
```
