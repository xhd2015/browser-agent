## Expected

Requirement **E4**:

- Exit 0; no error.
- Stdout non-empty; ends with `\n`.
- Mentions `chrome://extensions`, load unpacked, developer (case-insensitive).
- Mentions absolute install path when InstallPath populated.

## Side Effects

- Extension extracted under BaseDir.

## Errors

- Missing trailing newline glues shell prompt.

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
		t.Fatalf("install-cli error: %v", err)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	out := resp.Stdout
	assertContainsFold(t, out, "chrome://extensions")
	assertContainsFold(t, out, "load unpacked")
	assertContainsFold(t, out, "developer")

	if resp.InstallPath != "" {
		if !strings.Contains(out, resp.InstallPath) {
			t.Fatalf("stdout must include absolute install path %q; stdout=%q",
				resp.InstallPath, out)
		}
		mani := filepath.Join(resp.InstallPath, "manifest.json")
		if _, err := os.Stat(mani); err != nil {
			t.Fatalf("after install-cli, manifest missing at %s: %v", mani, err)
		}
	} else if !strings.Contains(out, req.BaseDir) && !strings.Contains(out, "/extension/") {
		t.Fatalf("stdout should mention extract path under BaseDir or /extension/; stdout=%q", out)
	}
}
```
