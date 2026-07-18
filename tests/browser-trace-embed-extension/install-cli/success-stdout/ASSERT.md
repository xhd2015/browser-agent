## Expected Output

User-facing install stdout is multi-line help. Exact layout is implementer-owned;
required tokens and **trailing newline** are asserted. When the path is known,
it must appear in stdout.

```
---
version: 3
__PATH__: type=string, example=/tmp/browser-trace-base/extension/1.2.0, absolute extract path
---
\.\.\.N lines omitted\.\.\.
```

(Strict full-template match is not used here because help copy may include
variable line counts; Assert uses token + trailing-`\n` checks instead.)

## Expected

Requirement scenario **#3** — `--install-chrome-extension` / `InstallChromeExtension`:

- Exit code 0; no error.
- Stdout is non-empty and **ends with `\n`**.
- Stdout mentions (case-insensitive where noted):
  - absolute install path (when Run fills `InstallPath`, that path must appear)
  - `chrome://extensions`
  - Load unpacked / `load unpacked`
  - Developer / `developer` (Developer mode step)
- Side effect: extract dir exists under BaseDir (manifest present when path known).

## Side Effects

- Extension extracted under `{BaseDir}/extension/{version}/`.

## Errors

- Must not exit non-zero on successful extract+print.
- Must not omit trailing newline (glues shell prompt).

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

	// Absolute path: prefer InstallPath from response; else scan for BaseDir/extension.
	if resp.InstallPath != "" {
		if !strings.Contains(out, resp.InstallPath) {
			t.Fatalf("stdout must include absolute install path %q; stdout=%q",
				resp.InstallPath, out)
		}
		mani := filepath.Join(resp.InstallPath, "manifest.json")
		if _, err := os.Stat(mani); err != nil {
			t.Fatalf("after install-cli, manifest missing at %s: %v", mani, err)
		}
	} else {
		// Fallback: stdout should still look like it contains an absolute path segment.
		if !strings.Contains(out, req.BaseDir) && !strings.Contains(out, "/extension/") {
			t.Fatalf("stdout should mention extract path under BaseDir or /extension/; stdout=%q", out)
		}
	}
}
```
