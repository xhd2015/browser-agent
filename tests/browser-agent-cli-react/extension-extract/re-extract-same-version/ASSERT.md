## Expected

Requirement **E2**:

- Both passes succeed.
- SecondPassInstallPath == InstallPath.
- SecondPassVersion == Version.
- Layout still valid.

## Side Effects

- Still only one version directory for that embed version.

## Errors

- Path drift is a hard fail.

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
}
```
