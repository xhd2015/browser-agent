## Expected

- Exit code **0**.
- Multiset equality:
  `sort(basenames under --out) == sort(browseragent.AssetReleaseNames(version))`.
- Helper returns the three known basenames for `v0.2.0`:
  - `browser-agent_v0.2.0_session-page.tar.gz`
  - `browser-agent_v0.2.0_extension.tar.gz`
  - `browser-trace_v0.2.0_extension.tar.gz`

## Side Effects

- Pack writes under temp `--out` only.

## Errors

- Non-zero exit or any basename mismatch fails.

## Exit Code

- 0.

```go
import (
	"sort"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertExitZero(t, resp)

	ver := req.Version
	if ver == "" {
		ver = ReleaseVersion
	}
	want := browseragent.AssetReleaseNames(ver)
	if len(want) == 0 {
		t.Fatalf("AssetReleaseNames(%q) returned empty", ver)
	}

	// Hard-pin known v0.2.0 contract so script cannot invent alternate schemes.
	if ver == "v0.2.0" || ver == "0.2.0" {
		for _, name := range []string{
			"browser-agent_v0.2.0_session-page.tar.gz",
			"browser-agent_v0.2.0_extension.tar.gz",
			"browser-trace_v0.2.0_extension.tar.gz",
		} {
			found := false
			for _, w := range want {
				if w == name {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("AssetReleaseNames missing %q; got=%v", name, want)
			}
		}
	}

	got := append([]string(nil), resp.OutBasenames...)
	sort.Strings(got)
	wantSorted := append([]string(nil), want...)
	sort.Strings(wantSorted)

	if len(got) != len(wantSorted) {
		t.Fatalf("basename count got=%d want=%d\n  got  %v\n  want %v\n  stdout=%s\n  stderr=%s",
			len(got), len(wantSorted), got, wantSorted,
			truncate(resp.Stdout, 300), truncate(resp.Stderr, 300))
	}
	for i := range wantSorted {
		if got[i] != wantSorted[i] {
			t.Fatalf("basename mismatch at %d:\n  got  %v\n  want %v", i, got, wantSorted)
		}
	}
	_ = strings.Join(got, ",")
}
```
