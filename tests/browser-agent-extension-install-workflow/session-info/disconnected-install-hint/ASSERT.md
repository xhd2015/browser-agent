## Expected Output

```text
  - Run: browser-agent install-chrome-extension
  - Load unpacked from: <path with extensions/browser-agent/>
```

## Expected

- CLI succeeds.
- Stdout contains `install-chrome-extension` and `extensions/browser-agent/`.
- Stdout does **not** suggest `browser-agent open-chrome`.

## Side Effects

- None.

## Errors

- Legacy open-chrome hint fails.

## Exit Code

- 0.

```go
import (
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
	assertContainsFold(t, resp.Stdout, "install-chrome-extension", "load unpacked", "extensions/browser-agent/")
	assertNotContainsFold(t, resp.Stdout, "browser-agent open-chrome")
}
```