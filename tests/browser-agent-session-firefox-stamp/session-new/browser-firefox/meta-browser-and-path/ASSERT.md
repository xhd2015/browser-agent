## Expected

- `SessionNew` returns nil error.
- `meta.json` exists under `{BaseDir}/sessions/{SessionID}/meta.json`.
- Meta stamps **browser** as firefox (field `browser` preferred; `browsers`
  containing firefox accepted).
- `extension_install_path` is non-empty and contains `browser-agent-firefox`.
- Path is **not** under `managed-chrome` and is not the Chrome-only
  `…/browser-agent/` tree.

## Side Effects

- Session directory created; extension may be extracted under TestHome.
- No browser open (NoOpenChrome).

## Errors

- Missing meta, chrome-stamped path, or absent browser firefox stamp fails.

## Exit Code

- N/A (package API).

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNewErr = %q", resp.SessionNewErr)
	}
	if strings.TrimSpace(resp.MetaJSON) == "" {
		t.Fatalf("meta.json missing or empty; path=%s", resp.MetaPath)
	}
	assertMetaBrowserFirefox(t, resp.MetaBrowser, resp.MetaJSON)
	assertFirefoxInstallPath(t, resp.MetaExtensionInstallPath)
	// Prefer absolute path when present.
	if p := resp.MetaExtensionInstallPath; p != "" && !filepath.IsAbs(p) {
		t.Fatalf("extension_install_path should be absolute; got %q", p)
	}
}
```
