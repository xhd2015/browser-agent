## Expected

Requirement **B4**:

- `JobResultOK` is false.
- Error text (Wait err and/or result.Error) contains `timeout` (case-insensitive).

## Side Effects

- Job may move to `expired` or `failed` (not strictly required here; see expire leaf).

## Errors

- ok=true without complete is a failure.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.JobResultOK {
		t.Fatal("JobResultOK=true, want false on wait timeout")
	}
	msg := strings.ToLower(resp.JobResultError)
	if !strings.Contains(msg, "timeout") {
		t.Fatalf("error should contain timeout; got %q", resp.JobResultError)
	}
}
```
