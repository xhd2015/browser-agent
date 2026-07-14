## Expected

Requirement scenario **#2** — re-extract same version:

- Both passes succeed without error.
- `SecondPassInstallPath` equals first `InstallPath` (stable path).
- `SecondPassVersion` equals first `Version`.
- Layout still valid after second pass (`manifest.json` readable; version match).

## Side Effects

- Still only `{BaseDir}/extension/{version}/` for that version (no duplicate version dirs).

## Errors

- Path drift between passes is a failure (breaks user-facing install instructions).

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("re-extract error: %v", err)
	}
	assertExitZero(t, resp)
	assertExtractLayout(t, req, resp)

	if resp.SecondPassInstallPath == "" {
		t.Fatal("SecondPassInstallPath empty; ExtractPasses=2 should populate it")
	}
	if resp.SecondPassInstallPath != resp.InstallPath {
		t.Fatalf("install path not stable: first=%q second=%q",
			resp.InstallPath, resp.SecondPassInstallPath)
	}
	if resp.SecondPassVersion != resp.Version {
		t.Fatalf("version not stable: first=%q second=%q",
			resp.Version, resp.SecondPassVersion)
	}
	// Re-validate layout after second pass (same path).
	assertExtractLayout(t, req, resp)
}
```
