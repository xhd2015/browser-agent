## Expected

After implementer lands session-rich (**RED** on current code):

- `GET /v1/session` returns HTTP 200.
- `session_page_count` is `0` (not omitted).
- `status` is `no_session_page`.
- `status_label` is present and human-readable (non-empty).

## Side Effects

- Fake extension hello with zero page count.

## Errors

- Missing status, wrong status, or nil/unknown page count fails.

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
	if !resp.WSHelloOK {
		t.Fatal("WSHelloOK=false; hello with page count 0 required")
	}
	assertHTTPStatus(t, resp, 200)

	if resp.SessionPageCount == nil {
		t.Fatalf("session_page_count missing; body=%s", truncate(resp.SessionBodyString, 600))
	}
	if *resp.SessionPageCount != 0 {
		t.Fatalf("session_page_count=%d want 0; body=%s", *resp.SessionPageCount, truncate(resp.SessionBodyString, 600))
	}
	if resp.Status != "no_session_page" {
		t.Fatalf("status=%q want no_session_page; body=%s", resp.Status, truncate(resp.SessionBodyString, 600))
	}
	if resp.StatusLabel == "" {
		t.Fatalf("status_label missing; body=%s", truncate(resp.SessionBodyString, 600))
	}
}
```