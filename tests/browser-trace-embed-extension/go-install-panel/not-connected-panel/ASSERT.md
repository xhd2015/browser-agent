## Expected

Requirement scenario **#6** — `/go` HTML while not connected:

- HTTP 200, HTML content type.
- Body non-empty; includes session id.
- Install panel marker present:
  - `data-browser-trace-install`, **or**
  - `id="browser-trace-install"` / `id='browser-trace-install'`, **or**
  - `id="browser-trace-install-panel"` / class containing `browser-trace-install`
- Body contains `chrome://extensions` as visible text (not only a dead chrome: link).
- Body contains absolute path guidance:
  - either `{BaseDir}/extension/` path segment, **or**
  - a `data-extension-path` / `data-install-path` attribute with an absolute path, **or**
  - the path string under `extension/` with a version segment
- Prefer Load unpacked / Developer mode wording somewhere in the panel.

## Side Effects

- Extract under BaseDir from session start.

## Errors

- Missing install panel when not connected is a failure.
- Relying only on `<a href="chrome://…">` without text is insufficient (blocked by Chrome).

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("probe transport error: %v", err)
	}
	assertHTTPStatus(t, resp, http.StatusOK)
	assertHTMLContentType(t, resp)

	body := resp.BodyString
	if body == "" {
		t.Fatal("HTML body is empty")
	}

	wantID := resp.RealSessionID
	if wantID == "" {
		wantID = req.SessionSuffix
	}
	if !strings.Contains(body, wantID) {
		t.Fatalf("HTML missing session id %q; body=%s", wantID, truncate(body, 400))
	}

	low := strings.ToLower(body)
	hasPanel := strings.Contains(low, "data-browser-trace-install") ||
		strings.Contains(low, `id="browser-trace-install"`) ||
		strings.Contains(low, `id='browser-trace-install'`) ||
		strings.Contains(low, `id="browser-trace-install-panel"`) ||
		strings.Contains(low, "browser-trace-install")
	if !hasPanel {
		t.Fatalf("HTML missing install panel marker (data-browser-trace-install or id=browser-trace-install); body=%s",
			truncate(body, 700))
	}

	if !strings.Contains(body, "chrome://extensions") {
		t.Fatalf("HTML must include chrome://extensions as text; body=%s", truncate(body, 600))
	}

	// Path guidance: BaseDir/extension, data attribute, or absolute .../extension/<ver>
	hasPath := strings.Contains(body, filepath.Join("extension", "")) ||
		strings.Contains(body, "/extension/") ||
		strings.Contains(body, req.BaseDir) ||
		strings.Contains(low, "data-extension-path") ||
		strings.Contains(low, "data-install-path")
	// Also accept any absolute-looking path with /extension/version pattern
	if !hasPath {
		re := regexp.MustCompile(`/extension/[^"'<\s]+`)
		if re.FindString(body) != "" {
			hasPath = true
		}
	}
	if !hasPath {
		t.Fatalf("HTML must show extension install path guidance; body=%s", truncate(body, 700))
	}

	// Prefer load/developer wording (soft-fail if panel marker + chrome:// already strong).
	if !strings.Contains(low, "load unpacked") && !strings.Contains(low, "developer") {
		t.Logf("warning: install panel missing Load unpacked / Developer wording; body snippet=%s",
			truncate(body, 300))
		// Require at least one of them for the contract.
		t.Fatalf("HTML install panel should mention Load unpacked and/or Developer mode; body=%s",
			truncate(body, 600))
	}
}
```
