## Expected

Requirement **S2**:

- SKILL.md exists.
- Body documents:
  - **`about:debugging`** (fragment URL optional)
  - temporary add-on / **Load Temporary Add-on** wording

## Side Effects

- None (read-only FS).

## Errors

- Missing about:debugging or temporary-add-on flow fails Firefox skill install docs.

## Exit Code

- Not asserted.

```go
import (
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
	if !hasAboutDebugging(body) {
		t.Fatalf("SKILL.md must document about:debugging; path=%v body=%s",
			resp.FoundPaths, truncate(body, 900))
	}
	if !hasTemporaryAddon(body) {
		t.Fatalf("SKILL.md must document temporary add-on / Load Temporary Add-on; path=%v body=%s",
			resp.FoundPaths, truncate(body, 900))
	}
}
```
