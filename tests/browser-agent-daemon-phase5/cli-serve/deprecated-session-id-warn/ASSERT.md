## Expected

- Stderr contains a **deprecation** token (case-insensitive).
- Stderr mentions **`--session-id`** or `session-id`.

## Side Effects

- Serve may block in background goroutine (not waited on).

## Errors

- Missing deprecation warning fails.

## Exit Code

- Not asserted (serve blocks).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.DeprecationWarnSeen {
		t.Fatalf("stderr missing deprecation warning; stderr=%q stdout=%q",
			truncate(resp.Stderr, 600), truncate(resp.Stdout, 200))
	}
	assertContainsFold(t, resp.Stderr, "--session-id", "session-id")
}
```