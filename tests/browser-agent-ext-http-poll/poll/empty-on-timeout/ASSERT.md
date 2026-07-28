## Expected

- HTTP status **200**.
- `jobs` empty (JobCount == 0).
- `events` empty (EventCount == 0).
- Response returns without client timeout (elapsed should be finite; soft check optional).

## Side Effects

- None required.

## Errors

- Non-200, non-empty jobs, or transport hang fails.

## Exit Code

- 0.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d want 200; body=%s",
			resp.StatusCode, truncate(resp.BodyString, 400))
	}
	if resp.JobCount != 0 {
		t.Fatalf("JobCount=%d want 0 on empty timeout; body=%s",
			resp.JobCount, truncate(resp.BodyString, 400))
	}
	if resp.EventCount != 0 {
		t.Fatalf("EventCount=%d want 0 on empty timeout; body=%s",
			resp.EventCount, truncate(resp.BodyString, 400))
	}
	// Soft: should not take far longer than wait_ms + generous slack.
	if resp.PollElapsedMS > 5000 {
		t.Fatalf("poll elapsed %dms; expected ~wait_ms=100 (possible hang)", resp.PollElapsedMS)
	}
	assertJSONContentType(t, resp.ContentType)
	assertExitZero(t, resp)
}
```
