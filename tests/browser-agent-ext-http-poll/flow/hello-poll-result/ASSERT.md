## Expected

- Hello left session **extension.connected=true**.
- Poll returned at least one job with type **`eval`** and non-empty id.
- Result HTTP **200** with `ok:true`.
- Background `POST /v1/jobs` completed with **ok=true**.

## Side Effects

- Full HTTP transport path works without WebSocket.

## Errors

- Any broken step (hello, poll empty, result missing, job hang) fails.

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
	if !resp.ExtensionConnected {
		t.Fatalf("expected extension.connected after hello; probe status=%d body/result=%s",
			resp.SessionProbeStatus, truncate(resp.BodyString, 400))
	}
	if resp.JobCount < 1 || resp.FirstJobID == "" {
		t.Fatalf("poll expected job; JobCount=%d FirstJobID=%q body=%s",
			resp.JobCount, resp.FirstJobID, truncate(resp.BodyString, 500))
	}
	if resp.FirstJobType != "eval" {
		t.Fatalf("job type=%q want eval", resp.FirstJobType)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("final result status=%d want 200; body=%s",
			resp.StatusCode, truncate(resp.BodyString, 400))
	}
	if !resp.ResultOKField {
		t.Fatalf("result ok missing/false; body=%s", truncate(resp.BodyString, 400))
	}
	if resp.HTTPJobError == "timeout waiting for /v1/jobs after flow result" {
		t.Fatalf("%s", resp.HTTPJobError)
	}
	if !resp.HTTPJobOK {
		t.Fatalf("POST /v1/jobs not ok after flow; err=%q body=%s",
			resp.HTTPJobError, truncate(resp.HTTPJobBody, 400))
	}
	assertExitZero(t, resp)
}
```
