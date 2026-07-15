## Expected

Requirement **S1**:

- SKILL.md exists and is non-empty (`browseragent/SKILL.md` preferred).
- Body documents **`create-tab`** (CLI) and/or job **`create_tab`**.
- Body mentions **Target** polyfill / polyfilled (or chrome.tabs tab lifecycle)
  **and** **`tab_id`** (or “tab id”) as identity.
- Soft: prefer create-tab / session info over raw target graph expectations.

## Side Effects

- Read-only filesystem read.

## Errors

- Missing create-tab or still Forbidden-only Target guidance without polyfill fails.

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

	if !strings.Contains(low, "create-tab") && !strings.Contains(body, "create_tab") {
		t.Fatalf("SKILL.md must document create-tab / create_tab; body=%s", truncate(body, 900))
	}

	// Explicit polyfill language required (legacy SKILL already pairs Forbidden Target + chrome.tabs).
	hasTarget := strings.Contains(body, "Target") || strings.Contains(low, "target.*")
	hasPolyfill := strings.Contains(low, "polyfill") || strings.Contains(low, "polyfilled")
	if !hasTarget || !hasPolyfill {
		t.Fatalf("SKILL.md must explicitly describe Target.* polyfill/polyfilled; body=%s",
			truncate(body, 900))
	}
	if !strings.Contains(body, "tab_id") && !strings.Contains(low, "tab id") {
		t.Fatalf("SKILL.md must mention tab_id identity; body=%s", truncate(body, 900))
	}
}
```
