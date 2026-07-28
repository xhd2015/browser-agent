## Expected

- Under `Firefox-Ext-Browser-Agent/public/`, all required files exist:
  - `manifest.json`
  - `background.js`
  - `contentScript.js`
  - `popup.html`
  - `popup.js`

## Side Effects

- None (read-only).

## Errors

- Any missing required file fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.RequiredFiles == nil {
		t.Fatal("RequiredFiles map is nil")
	}
	required := []string{"manifest.json", "background.js", "contentScript.js", "popup.html", "popup.js"}
	var missing []string
	for _, name := range required {
		if !resp.RequiredFiles[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("Firefox-Ext-Browser-Agent/public missing files %v under %s; found=%v",
			missing, resp.PackagePublicDir, resp.FoundPaths)
	}
}
```
