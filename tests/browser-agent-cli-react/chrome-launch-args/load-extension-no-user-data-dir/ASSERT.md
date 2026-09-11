## Expected

Requirement **F1** (default-profile session open):

- No error; ExitCode 0.
- ChromeArgs non-empty.
- Contains `--new-window`.
- Contains the request SessionURL.
- Does **not** contain `--load-extension` (joins Load-unpacked Chrome).
- Does **not** contain `--user-data-dir`.

## Side Effects

- May extract under BaseDir when ExtensionPath was empty.

## Errors

- Including `--user-data-dir` or `--load-extension` is a hard fail.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("chrome-args error: %v", err)
	}
	assertExitZero(t, resp)
	if resp.InstallPath == "" {
		t.Fatal("InstallPath empty; extract path still required for operator install hints")
	}
	assertChromeArgsContract(t, resp.ChromeArgs, resp.InstallPath, req.SessionURL)
}
```
