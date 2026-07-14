## Expected Output

```text
Extension:
  path    .../extensions/browser-agent/...
  install browser-agent install-chrome-extension

Note:
  Chrome 137+ cannot auto-load extensions.
```

## Expected

- Stdout contains `extensions/browser-agent/`, `install-chrome-extension`, `Chrome 137`.

## Side Effects

- Operator sees manual Load unpacked guidance.

## Errors

- Missing Extension block or Chrome 137 note fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	assertContainsFold(t, resp.Stdout,
		"extension:",
		"extensions/browser-agent/",
		"install-chrome-extension",
		"chrome 137",
		"load unpacked",
	)
}
```