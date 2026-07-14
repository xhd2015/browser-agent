## Expected Output

```
browser-agent
```

## Expected

Requirement **C1**:

- HandleCLI nil error; exit 0.
- Stdout contains skill name `browser-agent`.
- Stdout ends with trailing `\n`.
- Prefer first/only line exactly `browser-agent` (skillcmd Shape 1).

## Side Effects

- Read-only; no install.

## Errors

- Missing name / no trailing newline / hang fails.

## Exit Code

- 0 (CLI)

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertExitZero(t, resp)
	assertCLINilErr(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	if !strings.Contains(resp.Stdout, "browser-agent") {
		t.Fatalf("skill --list stdout must contain browser-agent; got %q", resp.Stdout)
	}
	// Prefer exact Shape-1 list output when no nested topics.
	trimmed := strings.TrimSuffix(resp.Stdout, "\n")
	if trimmed == "browser-agent" {
		assert.Output(t, resp.Stdout, `---
version: 2
---
browser-agent
`)
	} else {
		// Multi-line list still OK if first line is the skill name.
		first := strings.SplitN(trimmed, "\n", 2)[0]
		if first != "browser-agent" {
			t.Fatalf("skill --list first line = %q, want browser-agent; full=%q", first, resp.Stdout)
		}
	}
}
```
