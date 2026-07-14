## Expected

Requirement scenario **#1** — extract embedded extension to BaseDir:

- No error; exit code 0.
- `InstallPath` is absolute and lives under `{BaseDir}/extension/{version}`.
- `manifest.json` exists at `InstallPath/manifest.json`.
- Returned `Version` is non-empty and equals the manifest `"version"` field.
- Directory is readable (not empty of required files).

## Side Effects

- Creates `{BaseDir}/extension/{version}/` tree only (under BaseDir).

## Errors

- Must not return empty path/version on success.
- Must not write outside BaseDir.

## Exit Code

- 0.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	assertExitZero(t, resp)
	assertExtractLayout(t, req, resp)

	// Extra: at least one non-manifest file may exist for a realistic package,
	// but only manifest is strictly required for the mini fixture.
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
	_ = filepath.Separator
}
```
