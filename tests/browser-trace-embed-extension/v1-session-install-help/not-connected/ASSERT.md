## Expected

Requirement scenario **#5** — `/v1/session` while not connected (after extract):

- HTTP 200, JSON content type.
- Top-level `extension_install_path` is absolute and is a directory on disk.
- Top-level `embedded_version` is non-empty.
- `extension.connected` is false; `supports_browser_trace` is false.
- `hint` is non-empty and install-oriented: mentions at least one of
  load unpacked / chrome://extensions / install path / developer / extension install.
- Prefer: hint or body also contains the install path string or `chrome://extensions`.

## Side Effects

- Extract side effect under BaseDir from session start.

## Errors

- Missing `extension_install_path` after extract is a failure (session page cannot guide user).

## Exit Code

- Not asserted (probe-focused).

```go
import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("probe transport error: %v", err)
	}
	assertHTTPStatus(t, resp, http.StatusOK)
	assertJSONContentType(t, resp)

	assertInstallPathInJSON(t, resp)

	if resp.ExtensionConnected {
		t.Fatal("extension.connected = true, want false before hello")
	}
	if resp.SupportsBrowserTrace {
		t.Fatal("supports_browser_trace = true, want false before hello")
	}

	if strings.TrimSpace(resp.Hint) == "" {
		t.Fatalf("hint is empty; body=%s", truncate(resp.BodyString, 400))
	}
	h := strings.ToLower(resp.Hint)
	body := strings.ToLower(resp.BodyString)
	// Install guidance: path and/or chrome://extensions and/or load unpacked language.
	hasInstallLang := strings.Contains(h, "load unpacked") ||
		strings.Contains(h, "chrome://extensions") ||
		strings.Contains(h, "install") ||
		strings.Contains(h, "developer") ||
		strings.Contains(body, "chrome://extensions") ||
		strings.Contains(body, "load unpacked")
	if !hasInstallLang {
		t.Fatalf("hint/body should mention install guidance (Load unpacked / chrome://extensions / install); hint=%q body=%s",
			resp.Hint, truncate(resp.BodyString, 500))
	}

	// Stronger preference: path appears somewhere in the JSON payload.
	if !strings.Contains(resp.BodyString, resp.ExtensionInstallPath) {
		t.Fatalf("body should include extension_install_path value %q", resp.ExtensionInstallPath)
	}
	_ = filepath.Separator
}
```
