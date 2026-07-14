## Expected Output

```text
browser-agent: warning: Chrome 137+ ignores --load-extension; Load unpacked from <path>
```

## Expected

- CLI succeeds (managed launch still attempted).
- Stderr contains `Chrome 137`, `--load-extension`, and `Load unpacked`.

## Side Effects

- Operator warned about ignored load-extension flag.

## Errors

- Missing F1 warning fails.

## Exit Code

- 0 on successful dispatch.

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
	assertContainsFold(t, resp.Stderr,
		"warning",
		"chrome 137",
		"--load-extension",
		"load unpacked",
	)
}
```