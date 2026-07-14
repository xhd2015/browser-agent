## Expected

Requirement **E1**:

- No error; ExitCode 0.
- InstallPath absolute under `{BaseDir}/extension/{version}`.
- manifest.json exists; Version non-empty.

## Side Effects

- Creates extract tree only under BaseDir.

## Errors

- Empty path/version fails.

## Exit Code

- 0.

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	assertExitZero(t, resp)
	assertExtractLayout(t, req, resp)

	entries, err := os.ReadDir(resp.InstallPath)
	if err != nil {
		t.Fatalf("readdir install path: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("install path is empty")
	}
	foundManifest := false
	for _, e := range entries {
		if e.Name() == "manifest.json" {
			foundManifest = true
			break
		}
	}
	if !foundManifest {
		t.Fatalf("manifest.json not listed in %s", resp.InstallPath)
	}
}
```
