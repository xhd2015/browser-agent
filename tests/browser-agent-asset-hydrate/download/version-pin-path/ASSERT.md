## Expected

- At least one GET (`GETCount >= 1`).
- Combined request path/URL (slash path preferred) contains **`v0.2.0`**.
- Does **not** contain path segment / token **`latest`** as a release pin
  (case-sensitive `latest` in path).
- Path/URL includes product `browser-agent` and kind `session-page` (or
  archive name containing both).
- Prefer success path (`EnsureErr == nil`) so path inspection reflects a real
  ensure request; if Ensure fails but a GET was recorded with correct path,
  path pin asserts still apply — however GREEN implementer should succeed.

## Side Effects

- Cache may be filled under XDG when ensure succeeds.

## Errors

- No GET, missing version pin, or path contains `latest` fails.

## Exit Code

- 0.

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
	if resp.GETCount < 1 {
		t.Fatalf("GETCount=%d want >= 1 to inspect version-pinned path; ensureErr=%v",
			resp.GETCount, resp.EnsureErr)
	}
	path := resp.LastRequestPath
	if path == "" {
		path = resp.LastRequestURL
	}
	if path == "" {
		t.Fatal("no LastRequestPath/URL recorded")
	}
	if !strings.Contains(path, CacheVersion) && !strings.Contains(path, "v0.2.0") {
		t.Fatalf("request path %q missing version pin v0.2.0", path)
	}
	// Forbid "latest" as release selector in the path.
	if strings.Contains(path, "latest") {
		t.Fatalf("request path %q must not use latest; want version pin", path)
	}
	if !strings.Contains(path, ProductBrowserAgent) {
		t.Fatalf("request path %q missing product %q", path, ProductBrowserAgent)
	}
	if !strings.Contains(path, KindSessionPage) && !strings.Contains(path, "session-page") {
		t.Fatalf("request path %q missing kind session-page", path)
	}
	// Prefer ensure success for full GREEN contract.
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureAsset err=%v (path was %q); implementer should succeed with version-pinned URL",
			resp.EnsureErr, path)
	}
	assertExitZero(t, resp)
}
```
