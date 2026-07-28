## Expected

- `BuildExtensionShell` returns nil error.
- `ChromeBuildDir` / `BuildDir` is absolute.
- Path contains `Chrome-Ext-Browser-Agent/build`.
- `build/manifest.json` exists.
- Path does **not** contain `Firefox-Ext-Browser-Agent`.

## Side Effects

- Chrome build staged under ShellRoot only.

## Errors

- Wrong product shell or missing manifest fails.

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
	dir := resp.ChromeBuildDir
	if dir == "" {
		dir = resp.BuildDir
	}
	if strings.TrimSpace(dir) == "" {
		t.Fatal("Chrome BuildDir empty")
	}
	if !filepath.IsAbs(dir) {
		t.Fatalf("BuildDir should be absolute; got %q", dir)
	}
	norm := filepath.ToSlash(dir)
	if !strings.Contains(norm, "Chrome-Ext-Browser-Agent/build") {
		t.Fatalf("BuildExtensionShell should target Chrome-Ext-Browser-Agent/build; got %q", dir)
	}
	if strings.Contains(norm, "Firefox-Ext-Browser-Agent") {
		t.Fatalf("BuildExtensionShell must remain Chrome-only; got %q", dir)
	}
	mf := resp.BuildManifest
	if mf == "" {
		mf = filepath.Join(dir, "manifest.json")
	}
	if _, statErr := os.Stat(mf); statErr != nil {
		t.Fatalf("chrome build/manifest.json missing at %q: %v", mf, statErr)
	}
}
```
