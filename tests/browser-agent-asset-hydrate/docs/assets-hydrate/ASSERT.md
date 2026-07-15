## Expected

- At least one operator-facing markdown file is found under ModuleRoot
  (`docs/**/*.md`, `README.md`, `browseragent/SKILL.md`, …).
- Combined text (case-insensitive) mentions:
  1. **Cache path**: `~/.cache/browser-agent` **or** `asset-cache` **or**
     `.cache/browser-agent`
  2. **Incomplete embed / go install**: `go install` **or** `incomplete embed`
     **or** `go:embed` incomplete **or** `mini fixture`
  3. **CLI ensure**: `assets ensure` **or** (`assets` near `ensure`)
- Optional (not required): `BROWSER_AGENT_ASSET_BASE_URL`, `HTTPS_PROXY`.

## Side Effects

- None (read-only FS).

## Errors

- No docs found, or any of the three required topics missing, fails.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.DocsFound || strings.TrimSpace(resp.DocsCombinedText) == "" {
		t.Fatalf("no operator docs found under ModuleRoot=%s; expected docs/assets-hydrate.md and/or README/SKILL",
			req.ModuleRoot)
	}
	low := strings.ToLower(resp.DocsCombinedText)

	// 1) cache path
	hasCache := strings.Contains(low, "~/.cache/browser-agent") ||
		strings.Contains(low, ".cache/browser-agent") ||
		strings.Contains(low, "asset-cache") ||
		strings.Contains(low, "xdg_cache") && strings.Contains(low, "browser-agent")
	if !hasCache {
		t.Fatalf("docs missing cache path (~/.cache/browser-agent or asset-cache); paths=%v snippet=%s",
			resp.DocsPaths, truncate(resp.DocsCombinedText, 400))
	}

	// 2) go install / incomplete embed
	hasFallback := strings.Contains(low, "go install") ||
		strings.Contains(low, "incomplete embed") ||
		strings.Contains(low, "incomplete //go:embed") ||
		strings.Contains(low, "go:embed") && strings.Contains(low, "incomplete") ||
		strings.Contains(low, "mini fixture") ||
		strings.Contains(low, "fat release") && strings.Contains(low, "offline")
	if !hasFallback {
		t.Fatalf("docs missing go install / incomplete embed fallback; paths=%v snippet=%s",
			resp.DocsPaths, truncate(resp.DocsCombinedText, 400))
	}

	// 3) assets ensure
	hasEnsure := strings.Contains(low, "assets ensure") ||
		(strings.Contains(low, "assets") && strings.Contains(low, "ensure"))
	if !hasEnsure {
		t.Fatalf("docs missing assets ensure; paths=%v snippet=%s",
			resp.DocsPaths, truncate(resp.DocsCombinedText, 400))
	}

	assertExitZero(t, resp)
}
```
