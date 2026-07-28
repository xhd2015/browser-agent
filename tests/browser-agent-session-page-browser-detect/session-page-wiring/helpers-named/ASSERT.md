## Expected

Requirement **W1** (named pure helpers):

- Combined SessionPageApp / helper sources define **both**:
  - `detectRuntimeBrowser`
  - `resolveInstallBrowser`
- Prefer `export function` / `export const`, but non-exported `function` names
  are acceptable if clearly present.
- Soft: function-keyword or const arrow form extractable by name.

## Side Effects

- None (read-only FS).

## Errors

- Missing either helper name fails.

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

	if !hasDetectRuntimeBrowser(src) {
		t.Fatalf("must define detectRuntimeBrowser (Phase 1 pure UA helper); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}
	if !hasResolveInstallBrowser(src) {
		t.Fatalf("must define resolveInstallBrowser (Phase 1 priority helper); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	// Prefer real definitions, not comments only.
	detectDef := extractNamedFuncBody(src, "detectRuntimeBrowser")
	resolveDef := extractNamedFuncBody(src, "resolveInstallBrowser")
	if detectDef == "" {
		// Accept declaration-like patterns without full brace extract.
		if !(strings.Contains(src, "function detectRuntimeBrowser") ||
			strings.Contains(src, "detectRuntimeBrowser =") ||
			strings.Contains(src, "detectRuntimeBrowser(")) {
			t.Fatalf("detectRuntimeBrowser name present but no definition shape found; path=%v",
				resp.FoundPaths)
		}
	}
	if resolveDef == "" {
		if !(strings.Contains(src, "function resolveInstallBrowser") ||
			strings.Contains(src, "resolveInstallBrowser =") ||
			strings.Contains(src, "resolveInstallBrowser(")) {
			t.Fatalf("resolveInstallBrowser name present but no definition shape found; path=%v",
				resp.FoundPaths)
		}
	}
}
```
