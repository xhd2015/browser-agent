## Expected

- `HandleCLI --help` returns nil error.
- Help text lists `install-firefox-extension`.
- Help text still lists `install-chrome-extension` (Chrome path unchanged).
- Help ends with trailing `\n`.

## Side Effects

- None.

## Errors

- Missing new command or removed chrome command fails.

## Exit Code

- 0.

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
	if resp.CLIErr != "" {
		t.Fatalf("CLIErr = %q, want empty on --help", resp.CLIErr)
	}
	text := resp.Stdout + resp.Stderr
	if text == "" {
		t.Fatal("help output empty")
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("help must end with trailing newline")
	}
	assertContainsFold(t, text, "install-firefox-extension")
	assertContainsFold(t, text, "install-chrome-extension")
}
```
