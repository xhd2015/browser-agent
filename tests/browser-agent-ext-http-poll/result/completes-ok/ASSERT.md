## Expected

- Result HTTP status **200**.
- Result JSON `ok` is true (`ResultOKField`).
- Background `POST /v1/jobs` completes with **ok=true**.

## Side Effects

- Job reaches terminal done state; waiter unblocked.

## Errors

- Result 404, jobs still hanging, or jobs ok=false fails.

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
		t.Fatalf("result status=%d want 200; body=%s (route POST /v1/ext/result must exist)",
			resp.StatusCode, truncate(resp.BodyString, 500))
	}
	if !resp.ResultOKField {
		t.Fatalf("result ok field false or missing; body=%s", truncate(resp.BodyString, 400))
	}
	if resp.HTTPJobError == "timeout waiting for /v1/jobs after result" {
		t.Fatalf("POST /v1/jobs did not complete after result: %s", resp.HTTPJobError)
	}
	if !resp.HTTPJobOK {
		t.Fatalf("POST /v1/jobs ok=false after result; err=%q body=%s",
			resp.HTTPJobError, truncate(resp.HTTPJobBody, 400))
	}
	assertJSONContentType(t, resp.ContentType)
	assertExitZero(t, resp)
}
```
