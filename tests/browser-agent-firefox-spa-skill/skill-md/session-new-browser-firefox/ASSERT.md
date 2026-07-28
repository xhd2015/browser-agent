## Expected

Requirement **S3**:

- SKILL.md exists.
- Body documents Firefox session bootstrap via **`session new`** and
  **`--browser firefox`** (or equivalent clear recipe).
- Accept any of:
  - `session new --browser firefox`
  - `browser-agent session new --browser firefox`
  - `session new` … `--browser` … `firefox` in close proximity
- Must mention both `session new` (or `session-new`) and firefox browser flag.

## Side Effects

- None (read-only FS).

## Errors

- Chrome-only session new skill without Firefox browser flag fails S3.

## Exit Code

- Not asserted.

```go
import (
	"regexp"
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
	low := strings.ToLower(body)

	if !strings.Contains(low, "session new") && !strings.Contains(low, "session-new") {
		t.Fatalf("SKILL.md must mention session new; path=%v body=%s",
			resp.FoundPaths, truncate(body, 900))
	}

	// Preferred exact forms.
	if strings.Contains(body, "session new --browser firefox") ||
		strings.Contains(body, "session new --browser=firefox") ||
		strings.Contains(body, "session-new --browser firefox") {
		return
	}

	// Proximity: --browser and firefox near session new.
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)session\s+new[^.\n]{0,120}--browser[=\s]+firefox`),
		regexp.MustCompile(`(?is)--browser[=\s]+firefox[^.\n]{0,120}session\s+new`),
		regexp.MustCompile(`(?is)session\s+new[\s\S]{0,200}--browser[\s\S]{0,40}firefox`),
	}
	for _, re := range patterns {
		if re.MatchString(body) {
			return
		}
	}

	// Last resort: both tokens present with firefox browser guidance.
	if strings.Contains(body, "--browser") && strings.Contains(low, "firefox") &&
		(strings.Contains(low, "session new") || strings.Contains(low, "session-new")) {
		// Require firefox within 300 chars of session new.
		idx := strings.Index(low, "session new")
		if idx < 0 {
			idx = strings.Index(low, "session-new")
		}
		if idx >= 0 {
			end := idx + 300
			if end > len(low) {
				end = len(low)
			}
			window := low[idx:end]
			if strings.Contains(window, "firefox") && strings.Contains(window, "--browser") {
				return
			}
		}
	}

	t.Fatalf("SKILL.md must document session new --browser firefox; path=%v body=%s",
		resp.FoundPaths, truncate(body, 1000))
}
```
