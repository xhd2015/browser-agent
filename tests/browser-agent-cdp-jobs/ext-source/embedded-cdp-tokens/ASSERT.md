## Expected

Requirement **D3**:

- Embedded background found (package embed path preferred).
- Source contains `Runtime.evaluate`.
- Source mentions at least three of: `eval`, `run`, `logs`, `screenshot`, `cdp`, `info`
  (quoted preferred) OR contains `Page.captureScreenshot` / `chrome.debugger`
  in addition to Runtime.evaluate.
- Prefer full six types when cheap.

## Side Effects

- None (read-only; optional extract fallback).

## Errors

- Pure ok:true stub with zero CDP method names fails D3.

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
		t.Fatalf("embedded background missing; err=%q found=%v ModuleRoot=%s",
			resp.ErrText, resp.FoundPaths, req.ModuleRoot)
	}
	src := resp.CombinedText
	if !strings.Contains(src, "Runtime.evaluate") {
		t.Fatalf("embedded background missing Runtime.evaluate; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	types := []string{"eval", "run", "logs", "screenshot", "cdp", "info"}
	hit := 0
	for _, jt := range types {
		if strings.Contains(src, `"`+jt+`"`) || strings.Contains(src, `'`+jt+`'`) ||
			strings.Contains(src, jt) {
			// Count distinct type tokens with a light filter for eval/run/info.
			if jt == "eval" || jt == "run" || jt == "info" {
				if strings.Contains(src, `"`+jt+`"`) || strings.Contains(src, `'`+jt+`'`) ||
					strings.Contains(src, "type") && strings.Contains(src, jt) {
					hit++
				}
			} else {
				hit++
			}
		}
	}
	hasExtraCDP := strings.Contains(src, "Page.captureScreenshot") ||
		strings.Contains(src, "chrome.debugger") ||
		strings.Contains(src, "sendCommand")
	if hit < 3 && !hasExtraCDP {
		t.Fatalf("embedded background too thin: job-type hits=%d need≥3 or extra CDP tokens; path=%v snippet=%s",
			hit, resp.FoundPaths, truncate(src, 600))
	}
}
```
