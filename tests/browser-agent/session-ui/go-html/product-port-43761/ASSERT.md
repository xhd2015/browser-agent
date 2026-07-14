## Expected

Requirement **E4**:

- HTTP 200 HTML.
- Body contains `43761` (port in install text, boot config, or data attribute).
- Body mentions `browser-agent` (product display name / cli / config key),
  case-insensitive.

## Side Effects

- None.

## Errors

- Showing only 43759 (trace port) without 43761 fails product parameterization.

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
	body := resp.BodyString
	if !strings.Contains(body, "43761") {
		t.Fatalf("HTML/boot config must mention product port 43761; body=%s", truncate(body, 600))
	}
	if !strings.Contains(strings.ToLower(body), "browser-agent") {
		t.Fatalf("HTML/boot config must mention browser-agent product; body=%s", truncate(body, 600))
	}
}
```
