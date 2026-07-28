## Expected

- `SessionNew` returns nil error.
- `GET /v1/session` returns HTTP 200.
- `extension_install_path` (or bundled_extension.path) is non-empty and
  contains `browser-agent-firefox`.
- Path is not under `managed-chrome` / Chrome-only `browser-agent` tree.
- Optional: body may also expose browser/browsers firefox (soft if path ok).

## Side Effects

- In-memory session snap carries install path used by injectSessionBoot.

## Errors

- Non-200, empty path, or chrome-stamped path fails.

## Exit Code

- N/A.

```go
import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNewErr = %q", resp.SessionNewErr)
	}
	if resp.V1Status != http.StatusOK {
		t.Fatalf("GET /v1/session status=%d body=%s", resp.V1Status, truncate(resp.V1Body, 400))
	}
	p := resp.V1ExtensionInstallPath
	if strings.TrimSpace(p) == "" {
		t.Fatalf("v1 extension_install_path empty; body=%s", truncate(resp.V1Body, 500))
	}
	assertFirefoxInstallPath(t, p)
	if !filepath.IsAbs(p) {
		t.Fatalf("extension_install_path should be absolute; got %q", p)
	}
}
```
