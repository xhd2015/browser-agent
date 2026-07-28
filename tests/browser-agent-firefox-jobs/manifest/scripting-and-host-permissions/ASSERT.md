## Expected

- `manifest.json` found under `Firefox-Ext-Browser-Agent`.
- **Permissions**: includes **`tabs`** (already required for create/capture).
- **Script injection readiness** — at least one of:
  - `"scripting"` in `permissions`, **or**
  - background already on `tabs.executeScript` path is OK **if** hosts are broad
    (assert still prefers `"scripting"` when present; does not read background here).
  For this leaf: require **`"scripting"`** **or** document that implementer chose
  tabs.executeScript — we accept **`"scripting"`** **OR** presence of both broad
  hosts **and** `"tabs"` (tabs.executeScript fallback). Prefer requiring
  `"scripting"` **or** broad hosts with tabs.

- **host_permissions** not localhost-only: must include a broad pattern such as
  `<all_urls>`, `*://*/*`, `http://*/*`, `https://*/*`, or `*://*/*` family.

## Side Effects

- None.

## Errors

- Localhost-only host_permissions without broad hosts fails page inject/capture readiness.
- Missing tabs permission fails.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertManifestPresent(t, req, resp)
	low := strings.ToLower(src)

	if !strings.Contains(low, `"tabs"`) && !strings.Contains(low, "'tabs'") {
		// also accept unquoted rare forms
		if !strings.Contains(low, "tabs") {
			t.Fatalf("manifest must include tabs permission; path=%v snippet=%s",
				resp.FoundPaths, truncate(src, 500))
		}
	}

	hasScripting := strings.Contains(low, `"scripting"`) || strings.Contains(low, "'scripting'")
	hasBroadHosts := hasBroadHostPermissions(src)
	if !hasScripting && !hasBroadHosts {
		t.Fatalf("manifest needs scripting permission and/or broad host_permissions for page jobs; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	// Phase 2 page jobs: broad hosts required either way (localhost-only insufficient).
	if !hasBroadHosts {
		t.Fatalf("host_permissions must allow non-localhost page inject/capture (<all_urls> or http(s)://*/*); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
}

func hasBroadHostPermissions(src string) bool {
	low := strings.ToLower(src)
	patterns := []string{
		"<all_urls>",
		"*://*/*",
		"http://*/*",
		"https://*/*",
		"*://*/",
		"http://*/",
		"https://*/",
	}
	for _, p := range patterns {
		if strings.Contains(low, strings.ToLower(p)) {
			return true
		}
	}
	// Very broad match patterns sometimes written without quotes consistency.
	if strings.Contains(low, "all_urls") {
		return true
	}
	return false
}
```
