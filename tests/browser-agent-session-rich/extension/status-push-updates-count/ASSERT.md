## Expected

After implementer lands session-rich (**RED** on current code):

- `WSHelloOK` and `WSStatusSent` true.
- `GET /v1/session` HTTP 200 after status push.
- `session_page_count` is `2` (updated from initial `1`).
- `status` is `multiple_pages` after push.

## Side Effects

- Extension remains connected across hello and status.

## Errors

- Stale count=1 after status push fails.

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
		t.Fatal("WSHelloOK=false")
	}
	if !resp.WSStatusSent {
		t.Fatal("WSStatusSent=false")
	}
	assertHTTPStatus(t, resp, 200)

	if resp.SessionPageCount == nil || *resp.SessionPageCount != 2 {
		got := -1
		if resp.SessionPageCount != nil {
			got = *resp.SessionPageCount
		}
		t.Fatalf("session_page_count=%d want 2 after status push; body=%s",
			got, truncate(resp.SessionBodyString, 600))
	}
	if resp.Status != "multiple_pages" {
		t.Fatalf("status=%q want multiple_pages after count=2; body=%s",
			resp.Status, truncate(resp.SessionBodyString, 600))
	}
}
```