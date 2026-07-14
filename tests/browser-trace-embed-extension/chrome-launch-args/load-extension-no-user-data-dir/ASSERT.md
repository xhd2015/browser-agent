## Expected

Requirement scenario **#4** — Chrome launch arg builder:

- No error; exit code 0.
- `ChromeArgs` non-empty.
- Contains `--load-extension=<InstallPath>` (or `--load-extension` + path arg).
- Does **not** contain `--user-data-dir` or `--user-data-dir=…`.
- Contains the session URL used in the request.

## Side Effects

- May extract under BaseDir when ExtensionPath was empty (allowed).

## Errors

- Including `--user-data-dir` is a hard fail (isolated profile is a non-goal).

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
