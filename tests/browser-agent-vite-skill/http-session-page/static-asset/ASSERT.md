## Expected

Requirement **B1**:

- HTTP 200 for fixture asset path (default `/assets/session-page.js`).
- Body non-empty.
- Prefer Content-Type mentioning javascript / ecmascript / text / octet-stream
  (not strictly required if body non-empty and 200).

## Side Effects

- None.

## Errors

- 404 / empty body fails (fixture missing or not mounted).

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
	assertHTTPStatus(t, resp, http.StatusOK)
	if len(resp.Body) == 0 && strings.TrimSpace(resp.BodyString) == "" {
		t.Fatalf("static asset body empty; url=%q content-type=%q",
			resp.ProbeURL, resp.ContentType)
	}
	// Soft content-type check: warn-as-fail only if clearly HTML error page.
	ct := strings.ToLower(resp.ContentType)
	low := strings.ToLower(resp.BodyString)
	if strings.Contains(ct, "html") && strings.Contains(low, "<html") {
		t.Fatalf("static asset looks like HTML error page; url=%q ct=%q body=%s",
			resp.ProbeURL, resp.ContentType, truncate(resp.BodyString, 300))
	}
}
```
