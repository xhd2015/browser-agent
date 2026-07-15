## Expected

- `AssetReleaseNames("v0.2.0")` returns a non-empty slice.
- Slice contains exact basenames:
  - `browser-agent_v0.2.0_session-page.tar.gz`
  - `browser-agent_v0.2.0_extension.tar.gz`
- Recommended (not required for GREEN if agent-only): `browser-trace_v0.2.0_extension.tar.gz`.

## Side Effects

- None (pure).

## Errors

- Missing required agent archive names fails.

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
	if len(resp.ReleaseNames) == 0 {
		t.Fatal("AssetReleaseNames returned empty slice")
	}
	joined := "\n" + strings.Join(resp.ReleaseNames, "\n") + "\n"
	for _, want := range []string{ArchiveAgentSessionPage, ArchiveAgentExtension} {
		found := false
		for _, n := range resp.ReleaseNames {
			if n == want || strings.HasSuffix(n, "/"+want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("AssetReleaseNames missing %q; got=%v", want, resp.ReleaseNames)
		}
	}
	// Soft note: trace extension is recommended for dual-product releases.
	_ = ArchiveTraceExtension
	_ = joined
	assertExitZero(t, resp)
}
```
