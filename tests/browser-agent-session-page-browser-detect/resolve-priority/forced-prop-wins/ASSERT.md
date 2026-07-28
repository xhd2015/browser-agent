## Expected

Requirement **P1** (forced prop highest priority):

- Source defines **`resolveInstallBrowser`**.
- Function inspects a **forced** / prop argument early and returns when it is
  `"chrome"` or `"firefox"` (both branches).
- Forced check appears **before** content-script marker, UA, and snap/boot/path
  logic inside the function body (source order contract).

## Side Effects

- None (read-only FS).

## Errors

- Missing resolveInstallBrowser, missing forced early return, or forced after
  boot-only path fails.

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

	// Forced chrome|firefox early return signals.
	hasForcedChrome := strings.Contains(probe, `"chrome"`) || strings.Contains(probe, `'chrome'`)
	hasForcedFirefox := strings.Contains(probe, `"firefox"`) || strings.Contains(probe, `'firefox'`)
	if !hasForcedChrome || !hasForcedFirefox {
		t.Fatalf("resolveInstallBrowser must handle forced chrome and firefox; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 700))
	}

	// Heuristic: forced comparison near start of function.
	// Accept: forced === "firefox" || forced === "chrome", browserProp, first param checks.
	forcedSignals := []string{
		`forced === "firefox"`,
		`forced === "chrome"`,
		`forced==="firefox"`,
		`forced==="chrome"`,
		`forced === 'firefox'`,
		`forced === 'chrome'`,
		`"firefox" || forced === "chrome"`,
		`"chrome" || forced === "firefox"`,
		`forced === "firefox" || forced === "chrome"`,
		`forced === "chrome" || forced === "firefox"`,
	}
	foundForced := false
	for _, s := range forcedSignals {
		if strings.Contains(probe, s) {
			foundForced = true
			break
		}
	}
	// Broader: param named forced/browserProp compared to chrome|firefox with early return.
	if !foundForced {
		if (strings.Contains(probe, "forced") || strings.Contains(probe, "browserProp") || strings.Contains(probe, "browser:")) &&
			strings.Contains(probe, "return") &&
			(strings.Contains(probe, `"firefox"`) && strings.Contains(probe, `"chrome"`)) {
			foundForced = true
		}
	}
	if !foundForced {
		t.Fatalf("resolveInstallBrowser must early-return forced chrome|firefox prop; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 800))
	}

	// Forced should appear before snap/boot path markers when both present.
	idxForced := indexOfFirst(probe, "forced", "browserProp")
	idxBootish := indexOfFirst(probe,
		"__BROWSER_AGENT",
		"browser-agent-boot",
		"installPath",
		"snap?.browser",
		"snap.browser",
		"browsers",
	)
	if idxForced >= 0 && idxBootish >= 0 && idxForced > idxBootish {
		t.Fatalf("forced prop must be checked before snap/boot/path (idxForced=%d idxBoot=%d); body=%s",
			idxForced, idxBootish, truncate(probe, 800))
	}
}
```
