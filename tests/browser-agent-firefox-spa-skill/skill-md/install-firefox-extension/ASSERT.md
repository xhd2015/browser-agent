## Expected

Requirement **S1**:

- SKILL.md exists (`browseragent/SKILL.md` preferred; cmd path OK).
- Body documents **`install-firefox-extension`** (prefer
  `browser-agent install-firefox-extension`).

## Side Effects

- None (read-only FS).

## Errors

- Missing install-firefox-extension fails Firefox skill discoverability.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.SkillFileExists {
		t.Fatalf("SKILL.md not found; tried %v; err=%q", resp.SkillPathsTried, resp.ErrText)
	}
	body := resp.SkillText
	if !strings.Contains(body, "install-firefox-extension") {
		t.Fatalf("SKILL.md must document install-firefox-extension; path=%v body=%s",
			resp.FoundPaths, truncate(body, 900))
	}
}
```
