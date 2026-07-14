## Expected

Requirement **E3**:

- Extract succeeds; layout valid.
- ManifestText (or on-disk manifest) contains **`43761`**.
- Must not be only 43759 without 43761.

## Side Effects

- None beyond extract.

## Errors

- Missing 43761 means extension cannot reach control server.

## Exit Code

- 0.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	assertExitZero(t, resp)
	assertExtractLayout(t, req, resp)

	text := resp.ManifestText
	if strings.TrimSpace(text) == "" {
		mani := resp.ManifestPath
		if mani == "" {
			mani = filepath.Join(resp.InstallPath, "manifest.json")
		}
		data, rerr := os.ReadFile(mani)
		if rerr != nil {
			t.Fatalf("read manifest: %v", rerr)
		}
		text = string(data)
	}
	if !strings.Contains(text, "43761") {
		t.Fatalf("extracted manifest must mention port 43761; manifest=%s", truncate(text, 500))
	}
}
```
