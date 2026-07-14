## Expected

After implementer lands session-rich (**RED** on current code):

- `GET /v1/session` returns HTTP 200.
- Response JSON contains `created_at` with a non-empty value.
- `created_at` parses as a timestamp (RFC3339 or ISO8601).

## Side Effects

- Read-only probe after session create.

## Errors

- Missing or empty `created_at` fails.

## Exit Code

- Not asserted (HTTP probe only).

```go
import (
	"strings"
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertHTTPStatus(t, resp, 200)

	if resp.CreatedAt == "" {
		t.Fatalf("created_at missing from snapshot; body=%s", truncate(resp.SessionBodyString, 600))
	}
	if !strings.Contains(resp.SessionBodyString, "created_at") {
		t.Fatalf("raw JSON missing created_at key; body=%s", truncate(resp.SessionBodyString, 600))
	}
	if _, parseErr := time.Parse(time.RFC3339, resp.CreatedAt); parseErr != nil {
		if _, parseErr2 := time.Parse(time.RFC3339Nano, resp.CreatedAt); parseErr2 != nil {
			t.Fatalf("created_at=%q not RFC3339 parseable: %v / %v", resp.CreatedAt, parseErr, parseErr2)
		}
	}
}
```