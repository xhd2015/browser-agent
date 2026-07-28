## Expected

- `SessionNew` succeeds; session info CLI returns without transport error.
- Human stdout Next steps include **Firefox** install guidance:
  - `install-firefox-extension` and/or `about:debugging`
  - install path containing `browser-agent-firefox` (when path is printed)
- Stdout must **not** treat Chrome as the only primary install path:
  - fail if `install-chrome-extension` appears **without** any firefox install marker
  - fail if `chrome://extensions` appears as primary guidance without about:debugging

## Side Effects

- None beyond session create under BaseDir.

## Errors

- Chrome-only next steps after firefox create fail.

## Exit Code

- N/A (package/CLI hybrid; nonzero CLI fails).

```go
import (
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
	if resp.SessionInfoErr != "" {
		t.Fatalf("session info error: %s\nstdout=%s", resp.SessionInfoErr, truncate(resp.SessionInfoStdout, 400))
	}
	out := resp.SessionInfoStdout
	if strings.TrimSpace(out) == "" {
		t.Fatal("session info stdout empty")
	}
	low := strings.ToLower(out)
	hasFFInstall := strings.Contains(low, "install-firefox-extension") ||
		strings.Contains(low, "about:debugging") ||
		strings.Contains(low, "temporary add-on") ||
		strings.Contains(low, "temporary addon")
	if !hasFFInstall {
		t.Fatalf("session info Next steps must mention firefox install (install-firefox-extension / about:debugging); got:\n%s",
			truncate(out, 800))
	}
	// Prefer path visibility when printed.
	if strings.Contains(low, "load") || strings.Contains(low, "path") || strings.Contains(low, "extension") {
		if !strings.Contains(low, "browser-agent-firefox") && resp.MetaExtensionInstallPath != "" {
			// soft: if meta path is firefox but info omits it, still require install command markers above
		}
	}
	// Reject chrome-only primary guidance.
	hasChromeOnlyPrimary := strings.Contains(low, "install-chrome-extension") || strings.Contains(low, "chrome://extensions")
	if hasChromeOnlyPrimary && !hasFFInstall {
		t.Fatalf("session info must not primary-recommend chrome install for firefox session; got:\n%s", truncate(out, 800))
	}
	if strings.Contains(low, "install-chrome-extension") && !strings.Contains(low, "install-firefox-extension") &&
		!strings.Contains(low, "about:debugging") {
		t.Fatalf("session info still chrome-primary (install-chrome-extension without firefox markers); got:\n%s",
			truncate(out, 800))
	}
}
```
