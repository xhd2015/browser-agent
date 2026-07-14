## Expected

Requirement scenario **#1** — GET `/go` install re-run guidance:

- HTTP 200, HTML content type.
- Body non-empty.
- Mentions **close** (window / browser / Chrome).
- Mentions **re-run** or **run again** of **browser-trace**.
- Mentions install context: Load unpacked, Reload, or install/update.
- Optional but preferred: `data-install-rerun-guidance` attribute present.

## Side Effects

- Session dir may be created under BaseDir when Run starts; not asserted.

## Errors

- Missing close-window guidance is a failure.
- Missing re-run / browser-trace guidance is a failure.
- Relying only on existing Load unpacked path steps without re-run language is insufficient.

## Exit Code

- Not asserted (probe cancels Run early).

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
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /go status=%d, want 200; body=%q", resp.StatusCode, resp.BodyString)
	}
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "html") {
		t.Fatalf("Content-Type=%q, want text/html; body=%q", resp.ContentType, resp.BodyString)
	}
	body := resp.BodyString
	if strings.TrimSpace(body) == "" {
		t.Fatal("HTML body is empty")
	}
	low := strings.ToLower(body)

	// Close window / browser / Chrome.
	hasClose := strings.Contains(low, "close") &&
		(strings.Contains(low, "window") || strings.Contains(low, "chrome") || strings.Contains(low, "browser"))
	if !hasClose {
		t.Fatalf("HTML should tell user to close the Chrome/browser window; body snippet:\n%s", truncate(body, 800))
	}

	// Re-run browser-trace.
	hasRerun := (strings.Contains(low, "re-run") || strings.Contains(low, "rerun") ||
		strings.Contains(low, "run again") || strings.Contains(low, "run browser-trace again")) &&
		strings.Contains(low, "browser-trace")
	// Also accept "re-run" near browser-trace without exact phrase, or "run" + "browser-trace" + "again".
	if !hasRerun {
		if strings.Contains(low, "browser-trace") &&
			(strings.Contains(low, "again") || strings.Contains(low, "re-run") || strings.Contains(low, "rerun")) {
			hasRerun = true
		}
	}
	if !hasRerun {
		t.Fatalf("HTML should tell user to re-run browser-trace; body snippet:\n%s", truncate(body, 800))
	}

	// Install / Load unpacked / Reload context.
	hasInstallCtx := strings.Contains(low, "load unpacked") ||
		strings.Contains(low, "reload") ||
		strings.Contains(low, "install")
	if !hasInstallCtx {
		t.Fatalf("HTML should mention install/Load unpacked/Reload context; body snippet:\n%s", truncate(body, 800))
	}

	// Preferred stable marker (soft preference: log only if missing? — require for TDD clarity).
	if !strings.Contains(body, "data-install-rerun-guidance") {
		// Soft: do not fail if prose is complete; still encourage marker.
		// Requirement says optional — skip hard fail.
		_ = body
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
