# Expected

- HTTP **404**.
- Response body indicates the session was not found (case-insensitive match on
  substrings such as `not found` or `unknown` or `no such session`).
- Prefer JSON body; if Content-Type is JSON, structure may include `error` /
  `message` fields — either raw body text or those fields may satisfy the hint.

## Side Effects

- Live session is still torn down by harness cancel after the probe (no HAR
  assertions in this tree).

## Errors

- Must not return 200 with a fabricated empty session for an unknown id.
- Must not hang; probe times out in Run if the server never answers.

## Exit Code

- Not asserted (lifecycle exit after cancel is incidental).

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("probe transport error: %v", err)
	}
	assertHTTPStatus(t, resp, http.StatusNotFound)
	// Require JSON error body so a bare ServeMux "404 page not found" does not
	// false-green before GET /v1/session is implemented.
	assertJSONContentType(t, resp)

	body := strings.ToLower(resp.BodyString)
	// Also check common JSON error fields if parsed into Raw.
	if resp.Raw != nil {
		if m, _ := resp.Raw["error"].(string); m != "" {
			body += " " + strings.ToLower(m)
		}
		if m, _ := resp.Raw["message"].(string); m != "" {
			body += " " + strings.ToLower(m)
		}
		if m, _ := resp.Raw["hint"].(string); m != "" {
			body += " " + strings.ToLower(m)
		}
	}

	ok := strings.Contains(body, "not found") ||
		strings.Contains(body, "unknown") ||
		strings.Contains(body, "no such") ||
		strings.Contains(body, "does not exist") ||
		strings.Contains(body, "not exist")
	if !ok {
		t.Fatalf("404 body should indicate session not found; got: %s", truncate(resp.BodyString, 500))
	}
}
```
