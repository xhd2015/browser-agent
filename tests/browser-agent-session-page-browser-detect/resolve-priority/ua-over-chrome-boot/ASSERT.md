## Expected

Requirement **P3** (Firefox UA over chrome-stamped boot — primary Phase 1 bug):

- Source defines **`resolveInstallBrowser`**.
- Resolver consults **`detectRuntimeBrowser`** and/or reads **`userAgent`** /
  `navigator.userAgent` with a **Firefox/** detect.
- **Source order**: UA / `detectRuntimeBrowser` check appears **before**
  snap/boot/path chrome-stamped sources so a Firefox tab with chrome boot still
  resolves to firefox for InstallGuideline.
- Acceptable: call `detectRuntimeBrowser()` inside resolve, or inline the same
  UA check **after** forced + `__BROWSER_AGENT_EXT__` and **before** snap/boot.

## Side Effects

- None (read-only FS).

## Errors

- Snap/boot/path-only resolver without UA / detectRuntimeBrowser fails (current bug).

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertSourcePresent(t, req, resp, "SessionPageApp/helpers")

	if !hasResolveInstallBrowser(src) {
		t.Fatalf("resolveInstallBrowser must be defined; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	body := extractNamedFuncBody(src, "resolveInstallBrowser")
	probe := body
	if probe == "" {
		probe = src
	}

	// Must use detectRuntimeBrowser and/or UA Firefox detection inside resolve path.
	usesDetect := strings.Contains(probe, "detectRuntimeBrowser")
	usesUA := hasUserAgentRead(probe) && hasFirefoxUAToken(probe)
	if !usesDetect && !usesUA {
		// Also accept if resolve body calls detectRuntimeBrowser by name only
		// and detectRuntimeBrowser exists in session SPA text.
		spa := sessionSPA(resp)
		if strings.Contains(probe, "detectRuntimeBrowser") ||
			(hasDetectRuntimeBrowser(spa) && strings.Contains(probe, "detectRuntime")) {
			usesDetect = true
		}
	}
	if !usesDetect && !usesUA {
		t.Fatalf("resolveInstallBrowser must use detectRuntimeBrowser or UA Firefox before boot; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 900))
	}

	// Priority: UA/detect must come before snap/boot/path when those exist.
	idxUA := indexOfFirst(probe, "detectRuntimeBrowser", "userAgent", "UserAgent", "navigator", "Firefox/")
	idxSnapPath := indexOfFirst(probe,
		"installPath",
		"browser-agent-firefox",
		"browser-agent-boot",
		"snap?.browser",
		"snap.browser",
		"snap?.browsers",
	)
	// Prefer snap browsers array check only when clearly snap-driven (not forced).
	if idxSnapPath < 0 {
		idxSnapPath = indexOfFirst(probe, "browsers")
	}

	if idxUA < 0 {
		t.Fatalf("could not locate UA/detectRuntimeBrowser in resolveInstallBrowser; body=%s",
			truncate(probe, 700))
	}
	if idxSnapPath >= 0 && idxUA > idxSnapPath {
		t.Fatalf("UA/detectRuntimeBrowser must be checked before snap/boot/path (idxUA=%d idxSnap=%d); chrome-stamped boot must not win; body=%s",
			idxUA, idxSnapPath, truncate(probe, 1000))
	}

	// Soft: also before bare boot __BROWSER_AGENT (non-EXT) when present.
	// (EXT may correctly appear before UA — do not fail on that.)
}
```
