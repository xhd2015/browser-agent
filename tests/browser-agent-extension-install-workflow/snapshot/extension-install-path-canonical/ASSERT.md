## Expected

- HTTP 200 from `GET /v1/session`.
- `extension_install_path` non-empty.
- Path contains `extensions/browser-agent/` segment.
- Path is absolute.

## Side Effects

- Canonical extract during session create populates snapshot.

## Errors

- Empty path or legacy `{baseDir}/extension/` segment fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.HTTPStatus != http.StatusOK {
		t.Fatalf("GET /v1/session status=%d body=%s", resp.HTTPStatus, truncate(resp.BodyString, 400))
	}
	p := resp.SessionJSONExtensionInstallPath
	if strings.TrimSpace(p) == "" {
		t.Fatalf("extension_install_path empty; body=%s", truncate(resp.BodyString, 400))
	}
	assertCanonicalPathSegment(t, p)
	if !filepath.IsAbs(p) {
		t.Fatalf("extension_install_path should be absolute; got %q", p)
	}
}
```