## Expected

Requirement **E5** (light):

- background.js found.
- At least **one** of:
  - Soft Tier B tokens: `setDiscoverTargets`, `setAutoAttach`, `attachToTarget`,
    `detachFromTarget`
  - Unsupported / polyfill error language: `unsupported`, `polyfill`,
    `not supported`, `polyfill unsupported`
- Soft: prefer that unsupported path is product-owned (mention polyfill /
  unsupported) rather than only rethrowing Chrome `-32000` / `Not allowed`.

## Side Effects

- None.

## Errors

- No soft methods and no unsupported-handler language is incomplete polyfill surface.

## Exit Code

- Not asserted.

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
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("shell background missing; err=%q found=%v", resp.ErrText, resp.FoundPaths)
	}
	src := resp.CombinedText
	low := strings.ToLower(src)

	soft := []string{
		"setDiscoverTargets",
		"setAutoAttach",
		"attachToTarget",
		"detachFromTarget",
		"Target.setDiscoverTargets",
		"Target.setAutoAttach",
		"Target.attachToTarget",
		"Target.detachFromTarget",
	}
	softHit := false
	for _, s := range soft {
		if strings.Contains(src, s) {
			softHit = true
			break
		}
	}

	unsupHit := (strings.Contains(low, "polyfill") &&
		(strings.Contains(low, "unsupported") || strings.Contains(low, "not supported") ||
			strings.Contains(low, "not implemented"))) ||
		strings.Contains(low, "polyfill unsupported") ||
		strings.Contains(low, "unsupported target") ||
		strings.Contains(src, "TARGET_POLYFILL") ||
		strings.Contains(src, "unsupportedTarget")

	if !softHit && !unsupHit {
		t.Fatalf("background need Tier B soft Target methods and/or Tier C polyfill-unsupported errors; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
