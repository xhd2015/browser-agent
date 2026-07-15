## Expected

- `docs/assets-hydrate.md` exists under ModuleRoot and is non-empty.
- Body contains **`script/github/release-assets`** (path token; `go run ./script/github/release-assets` also satisfies via substring).
- Body contains **`--upload`** (exact flag form).
- Optional (not required for GREEN): mention of `gh`, `clobber`, pack, or `AssetReleaseNames`.
- Prefer real path tokens; body must not rely solely on dotted scaffold placeholder IDs
  (e.g. bare `P6.1`-style IDs without the script path). This leaf only hard-requires
  the two tokens above.

## Side Effects

- None (read-only FS).

## Errors

- Missing file, empty body, or missing either required token fails.

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
	if !resp.DocsOK || strings.TrimSpace(resp.DocsText) == "" {
		t.Fatalf("docs/assets-hydrate.md missing or empty; path=%q err=%q",
			resp.DocsPath, resp.ErrText)
	}

	text := resp.DocsText
	// Path token (also matches "go run ./script/github/release-assets").
	if !strings.Contains(text, "script/github/release-assets") {
		t.Fatalf("docs/assets-hydrate.md missing script/github/release-assets; path=%s snippet=%s",
			resp.DocsPath, truncate(text, 500))
	}
	if !strings.Contains(text, "--upload") {
		t.Fatalf("docs/assets-hydrate.md missing --upload; path=%s snippet=%s",
			resp.DocsPath, truncate(text, 500))
	}

	assertExitZero(t, resp)
}
```
