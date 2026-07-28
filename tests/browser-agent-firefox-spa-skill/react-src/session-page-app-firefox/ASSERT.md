## Expected

Requirement **R2** (SessionPageApp Firefox wiring):

- `SessionPageApp` source exists (prefer `react/src/ui/SessionPageApp.tsx`).
- Combined SessionPageApp (+ InstallGuideline when augmented) shows Firefox
  install awareness via **at least two** of:
  - `about:debugging`
  - temporary add-on / Load Temporary Add-on
  - `browser-agent-firefox`
  - browser signal: `"firefox"` / `'firefox'` / `browser === "firefox"` /
    `browsers` + firefox / `install-firefox` / `isFirefox` / `browserKind`
- Prefer explicit wiring: pass a browser prop into `InstallGuideline`, or
  conditional Firefox steps in SessionPageApp itself.
- Soft: still importing/using `InstallGuideline` for the not-connected panel.

## Side Effects

- None (read-only FS).

## Errors

- Chrome-only install copy with zero Firefox markers / browser signal fails.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertFilePresent(t, req, resp, "SessionPageApp")

	// SessionPageApp should still host install UX when not connected.
	if !strings.Contains(src, "InstallGuideline") &&
		!strings.Contains(src, "install-guideline") &&
		!strings.Contains(src, "about:debugging") {
		t.Fatalf("SessionPageApp should use InstallGuideline or inline firefox install steps; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}

	score := 0
	if hasAboutDebugging(src) {
		score++
	}
	if hasTemporaryAddon(src) {
		score++
	}
	if hasFirefoxExtPath(src) {
		score++
	}
	// Browser detection / prop wiring signals.
	browserSignal := strings.Contains(src, `"firefox"`) ||
		strings.Contains(src, `'firefox'`) ||
		strings.Contains(src, "`firefox`") ||
		strings.Contains(src, "browser === \"firefox\"") ||
		strings.Contains(src, "browser==='firefox'") ||
		strings.Contains(src, "isFirefox") ||
		strings.Contains(src, "browserKind") ||
		strings.Contains(src, "install-firefox") ||
		strings.Contains(src, "installFirefox") ||
		(strings.Contains(src, "browsers") && strings.Contains(strings.ToLower(src), "firefox"))
	if browserSignal {
		score++
	}

	if score < 2 {
		t.Fatalf("SessionPageApp Firefox wiring weak (score=%d want >=2 of about:debugging / temporary-addon / browser-agent-firefox / browser-signal); path=%v snippet=%s",
			score, resp.FoundPaths, truncate(src, 800))
	}
}
```
