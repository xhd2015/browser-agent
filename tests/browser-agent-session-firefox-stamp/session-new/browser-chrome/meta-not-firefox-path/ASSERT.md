## Expected

- `SessionNew` returns nil error.
- `meta.json` has non-empty `extension_install_path`.
- Path is the **Chrome** install tree (contains `browser-agent`, does **not**
  contain `browser-agent-firefox`).
- If `browser` field is present, it must not be forced to firefox.

## Side Effects

- Chrome extension may extract under process or package home.

## Errors

- Firefox path on chrome create fails.

## Exit Code

- N/A.

```go
import (
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
	assertChromeInstallPath(t, resp.MetaExtensionInstallPath)
	if strings.EqualFold(strings.TrimSpace(resp.MetaBrowser), "firefox") {
		t.Fatalf("chrome create must not stamp browser=firefox; meta=%s", truncate(resp.MetaJSON, 400))
	}
}
```
