## Expected

Requirement **B2**:

- Non-empty system prompt.
- Body does **not** contain the concrete control id `ctrl-sysmd-unique-qq7`.

## Side Effects

- None (pure).

## Errors

- Embedding control id in SYSTEM.md couples playbook copies to one session.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	p := resp.SystemPrompt
	if strings.TrimSpace(p) == "" {
		t.Fatal("SystemPrompt empty")
	}
	ctrl := req.PromptSessionID
	if strings.Contains(p, ctrl) {
		t.Fatalf("SYSTEM.md body must not embed control session id %q; prompt=%s",
			ctrl, truncate(p, 600))
	}
}
```
