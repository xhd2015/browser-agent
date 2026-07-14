## Expected

- Nil error from Run.
- ChromeArgs satisfy managed contract with empty url.
- Contains `--user-data-dir` pointing at `Layout.DataDir`.
- Contains `--load-extension` for extracted extension path.
- Contains `--new-window`.
- No `http://` or `https://` argument.

## Side Effects

- Extension extracted under ManagedRoot.

## Errors

- URL present in blank-window mode fails.

## Exit Code

- N/A.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("chrome-args blank-window error: %v", err)
	}
	assertManagedChromeArgsContract(t, resp.ChromeArgs, resp.Layout.DataDir, resp.ExtensionPath, "")
}```
