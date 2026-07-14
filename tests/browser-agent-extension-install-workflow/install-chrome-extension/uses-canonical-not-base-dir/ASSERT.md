## Expected

- CLI succeeds.
- Stdout path contains `extensions/browser-agent/`.
- Stdout path does **not** contain `{BaseDir}/extension/`.

## Side Effects

- Extract under home canonical layout only.

## Errors

- Path rooted under custom `--base-dir` fails.

## Exit Code

- 0 on success.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.CLIErr != "" {
		t.Fatalf("CLI error: %s", resp.CLIErr)
	}
	assertCanonicalPathSegment(t, resp.Stdout)
	legacy := filepath.ToSlash(filepath.Join(req.BaseDir, "extension"))
	if strings.Contains(filepath.ToSlash(resp.Stdout), legacy) {
		t.Fatalf("stdout should not use base-dir extension path %q; got:\n%s", legacy, resp.Stdout)
	}
}
```