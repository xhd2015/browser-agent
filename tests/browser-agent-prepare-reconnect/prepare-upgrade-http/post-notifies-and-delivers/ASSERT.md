## Expected

- HTTP status **200**.
- JSON `ok` is true.
- `notified` contains **`sess-admin-a`**.
- Connected WS receives `type=prepare_reconnect` with `delay_ms=1000`.

## Side Effects

- Extension schedules reconnect after delay (not executed in this fake client).

## Errors

- 404 (route missing), empty notified, or no WS delivery fails.

## Exit Code

- 0.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d want 200; body=%s (route /v1/admin/prepare-upgrade must exist)",
			resp.StatusCode, resp.BodyString)
	}
	if !resp.HTTPOK {
		t.Fatalf("ok field false or missing; body=%s", resp.BodyString)
	}
	assertNotifiedEquals(t, resp.HTTPNotifiedIDs, []string{"sess-admin-a"})
	if !resp.HTTPWSReceived {
		t.Fatalf("WS prepare_reconnect not received after POST; body=%s", resp.BodyString)
	}
	if resp.HTTPWSType != "prepare_reconnect" {
		t.Fatalf("WS type=%q want prepare_reconnect", resp.HTTPWSType)
	}
	if resp.HTTPWSDelayMS != 1000 {
		t.Fatalf("WS delay_ms=%d want 1000", resp.HTTPWSDelayMS)
	}
	if ct := strings.ToLower(resp.ContentType); ct != "" && !strings.Contains(ct, "json") {
		t.Fatalf("Content-Type=%q want application/json", resp.ContentType)
	}
	assertExitZero(t, resp)
}
```
