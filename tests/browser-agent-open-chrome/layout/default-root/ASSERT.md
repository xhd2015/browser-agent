## Expected

- `DefaultManagedChromeLayout` returns nil error.
- `Layout.Root` is absolute and ends with `.browser-agent/managed-chrome` (platform path sep).
- `Layout.DataDir` = `{Root}/data`.
- `Layout.ExtensionsDir` = `{Root}/extensions`.

## Side Effects

- None (read-only path resolution).

## Errors

- Resolution error or wrong subdirs fails.

## Exit Code

- N/A (package API).

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("DefaultManagedChromeLayout error: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	wantRoot := filepath.Join(home, ".browser-agent", "managed-chrome")
	if !filepath.IsAbs(resp.Layout.Root) {
		t.Fatalf("Layout.Root %q is not absolute", resp.Layout.Root)
	}
	// Normalize for comparison.
	gotRoot, _ := filepath.Abs(resp.Layout.Root)
	wantAbs, _ := filepath.Abs(wantRoot)
	if gotRoot != wantAbs {
		t.Fatalf("Layout.Root = %q, want %q", gotRoot, wantAbs)
	}
	if !strings.HasSuffix(gotRoot, filepath.Join(".browser-agent", "managed-chrome")) {
		t.Fatalf("Layout.Root should end with .browser-agent/managed-chrome; got %q", gotRoot)
	}
	assertLayoutUnderRoot(t, resp.Layout, gotRoot)
}```
