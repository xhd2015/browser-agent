## Expected

Requirement scenario **#7** — connected + supports browser-trace:

- HTTP 200 JSON.
- `extension.connected` true; `supports_browser_trace` true.
- `hint` is non-empty and **not** primarily an install tutorial:
  - must **not** be solely "Load unpacked" instructions as the main message
  - should mention connected / supports / recording / browse (operational language)
- `extension_install_path` / `embedded_version` **may** still be present (allowed);
  presence is not required for this leaf if product hides them when supported.

## Side Effects

- None beyond short-lived session.

## Errors

- If supports=true but hint still only tells user to Load unpacked, demotion failed.

## Exit Code

- Not asserted.

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
	assertHTTPStatus(t, resp, http.StatusOK)
	assertJSONContentType(t, resp)

	if !resp.ExtensionConnected {
		t.Fatalf("extension.connected = false, want true after capable hello; body=%s",
			truncate(resp.BodyString, 400))
	}
	if !resp.SupportsBrowserTrace {
		t.Fatalf("supports_browser_trace = false, want true; body=%s",
			truncate(resp.BodyString, 400))
	}

	if strings.TrimSpace(resp.Hint) == "" {
		t.Fatal("hint is empty after capable hello")
	}
	h := strings.ToLower(resp.Hint)

	// Operational language preferred.
	operational := strings.Contains(h, "connect") ||
		strings.Contains(h, "support") ||
		strings.Contains(h, "record") ||
		strings.Contains(h, "brows") ||
		strings.Contains(h, "captur") ||
		strings.Contains(h, "ready") ||
		strings.Contains(h, "window")
	if !operational {
		t.Fatalf("hint should be operational (connected/supports/recording), not install-only; hint=%q",
			resp.Hint)
	}

	// Primary install tutorial should not dominate when support is OK.
	// Fail if hint looks like a pure install checklist without operational context.
	installHeavy := strings.Contains(h, "load unpacked") &&
		strings.Contains(h, "chrome://extensions") &&
		!operational
	if installHeavy {
		t.Fatalf("hint still primary install tutorial while supports=true; hint=%q", resp.Hint)
	}
}
```
