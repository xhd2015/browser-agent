## Expected

- `cmd/browser-agent/SKILL.md` exists and is non-empty.
- Body documents **`session new`** bootstrap (or equivalent “create session” recipe).
- Body documents **`serve --status`** read-only probe.
- Body documents post-bootstrap **`session info`** / **`session eval`** workflow.
- Body marks **`serve --session-id`** as **deprecated** (or directs to `session new` /
  plain `serve` instead of primary `serve --session-id` bootstrap).

## Side Effects

- Read-only filesystem read.

## Errors

- Missing workflow markers or missing deprecation note fails.

## Exit Code

- N/A (filesystem probe).

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
	if !resp.SkillFileExists {
		t.Fatalf("SKILL.md not found; tried %v; err=%q", resp.SkillPathsTried, resp.ErrText)
	}
	body := resp.SkillText
	low := strings.ToLower(body)
	assertContainsFold(t, body, "session new", "serve --status")
	for _, cmd := range []string{"session info", "session eval"} {
		if !strings.Contains(low, cmd) {
			t.Fatalf("SKILL.md must document %q workflow; body=%s", cmd, truncate(body, 900))
		}
	}
	deprecatedOK := strings.Contains(low, "deprecat") && strings.Contains(low, "serve --session-id")
	primaryBootstrapOK := strings.Contains(low, "browser-agent serve\n") ||
		strings.Contains(low, "browser-agent session new")
	if !deprecatedOK && !primaryBootstrapOK {
		t.Fatalf("SKILL.md must deprecate serve --session-id or document serve|session new bootstrap; body=%s",
			truncate(body, 900))
	}
	if strings.Contains(body, "browser-agent serve --session-id <id>") && !strings.Contains(low, "deprecat") {
		t.Fatalf("SKILL.md still recommends serve --session-id without deprecation; body=%s",
			truncate(body, 900))
	}
}
```