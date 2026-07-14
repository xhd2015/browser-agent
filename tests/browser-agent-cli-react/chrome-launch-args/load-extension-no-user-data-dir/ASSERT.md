## Expected

Requirement **F1**:

- No error; ExitCode 0.
- ChromeArgs non-empty.
- Contains `--load-extension=<InstallPath>` (or split form).
- Does **not** contain `--user-data-dir`.
- Contains the request SessionURL.

## Side Effects

- May extract under BaseDir when ExtensionPath was empty.

## Errors

- Including `--user-data-dir` is a hard fail.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("chrome-args error: %v", err)
	}
	assertExitZero(t, resp)
	if resp.InstallPath == "" {
		t.Fatal("InstallPath empty; need extract path for --load-extension assert")
	}
	assertChromeArgsContract(t, resp.ChromeArgs, resp.InstallPath, req.SessionURL)
}
```
