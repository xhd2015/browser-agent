## Expected

- HTTP status **200** with job result body (not 404).
- `ok` is **false**.
- Error mentions not connected and/or timeout (not silent success).
- `data.hint` mentions `/go?session=sess-job` and keep-open / do-not-navigate guidance.
- `data.session_url` present when possible; must contain `/go?session=sess-job` if set.

## Side Effects

- Job does not succeed without extension.

## Errors

- `ok:true` without extension fails.
- Missing `data.hint` fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.StatusCode == http.StatusNotFound {
		t.Fatalf("got 404 for known session; body=%s", truncate(resp.BodyString, 300))
	}
	if resp.HTTPJobOK {
		t.Fatal("HTTPJobOK=true without extension, want failure")
	}
	msg := strings.ToLower(resp.HTTPJobError + " " + resp.BodyString)
	if !strings.Contains(msg, "connect") && !strings.Contains(msg, "timeout") && !strings.Contains(msg, "extension") {
		t.Fatalf("expected not-connected or timeout signal; err=%q body=%s",
			resp.HTTPJobError, truncate(resp.BodyString, 400))
	}
	if strings.TrimSpace(resp.HTTPJobDataHint) == "" {
		t.Fatalf("data.hint missing; body=%s", truncate(resp.BodyString, 500))
	}
	assertDisconnectedHint(t, resp.HTTPJobDataHint, req.SessionID)
	if resp.HTTPJobDataSessionURL != "" {
		if !strings.Contains(resp.HTTPJobDataSessionURL, "/go?session="+req.SessionID) {
			t.Fatalf("data.session_url=%q want /go?session=%s", resp.HTTPJobDataSessionURL, req.SessionID)
		}
	}
}
```