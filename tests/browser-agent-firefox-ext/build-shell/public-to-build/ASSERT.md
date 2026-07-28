## Expected

- `BuildFirefoxExtensionShell` returns nil error.
- `BuildDir` is absolute and non-empty.
- `build/manifest.json` exists at `BuildManifest`.
- Build path is under `Firefox-Ext-Browser-Agent/build` (segment check).

## Side Effects

- Files written under ShellRoot/Firefox-Ext-Browser-Agent/build/.

## Errors

- Missing API or copy failure fails.

## Exit Code

- Not asserted.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if strings.TrimSpace(resp.BuildDir) == "" {
		t.Fatal("BuildDir empty")
	}
	if !filepath.IsAbs(resp.BuildDir) {
		t.Fatalf("BuildDir should be absolute; got %q", resp.BuildDir)
	}
	norm := filepath.ToSlash(resp.BuildDir)
	if !strings.Contains(norm, "Firefox-Ext-Browser-Agent/build") {
		t.Fatalf("BuildDir should contain Firefox-Ext-Browser-Agent/build; got %q", resp.BuildDir)
	}
	if _, statErr := os.Stat(resp.BuildManifest); statErr != nil {
		t.Fatalf("build/manifest.json missing at %q: %v", resp.BuildManifest, statErr)
	}
}
```
