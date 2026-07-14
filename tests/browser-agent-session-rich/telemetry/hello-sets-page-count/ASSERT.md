## Expected

After implementer lands session-rich (**RED** on current code):

- `WSHelloOK` true.
- `GET /v1/session` HTTP 200.
- `session_page_count` is `1`.
- `browsers` includes `Chrome` (or `browser_product` reflected in `browsers`).

## Side Effects

- Extension connected via hello.

## Errors

- Missing page count or browser product fails.

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
	if !resp.WSHelloOK {
		t.Fatal("WSHelloOK=false")
	}
	assertHTTPStatus(t, resp, 200)

	if resp.SessionPageCount == nil || *resp.SessionPageCount != 1 {
		got := -1
		if resp.SessionPageCount != nil {
			got = *resp.SessionPageCount
		}
		t.Fatalf("session_page_count=%d want 1; body=%s", got, truncate(resp.SessionBodyString, 600))
	}

	hasChrome := false
	for _, b := range resp.Browsers {
		if strings.EqualFold(b, "Chrome") {
			hasChrome = true
			break
		}
	}
	if !hasChrome {
		low := strings.ToLower(resp.SessionBodyString)
		if !strings.Contains(low, "chrome") {
			t.Fatalf("browsers missing Chrome; browsers=%v body=%s", resp.Browsers, truncate(resp.SessionBodyString, 600))
		}
	}
}
```