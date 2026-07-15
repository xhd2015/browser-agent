## Expected

Requirement **C2**:

- Non-empty system prompt.
- Mentions **Target** together with **polyfill** / **polyfilled** (or equivalent
  “implemented via chrome.tabs” language).
- Mentions **`tab_id`** as public identity for tab lifecycle / Target results.
- Prefer create-tab / session info for tab lifecycle over inventing target graphs.
- **Must not** be Forbidden-only: if the word `Forbidden` still appears near
  Target, the prompt must **also** describe polyfill/polyfilled path.
  Acceptable: polyfill language present even if some Target methods remain limited.

## Side Effects

- None.

## Errors

- Leaving only “Forbidden Target.* / -32000 Not allowed” without polyfill guidance fails.

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
	low := strings.ToLower(p)

	// Require explicit polyfill language — co-mention of chrome.tabs + Target is NOT enough
	// (legacy Forbidden Target.* docs already say use session info / chrome.tabs).
	hasPolyfillWord := strings.Contains(low, "polyfill") || strings.Contains(low, "polyfilled")
	hasTarget := strings.Contains(p, "Target") || strings.Contains(low, "target.*")
	if !hasTarget || !hasPolyfillWord {
		t.Fatalf("prompt must explicitly describe Target.* as polyfill/polyfilled; prompt=%s",
			truncate(p, 800))
	}

	if !strings.Contains(p, "tab_id") && !strings.Contains(low, "tab id") {
		t.Fatalf("prompt must mention tab_id identity for Target/create-tab results; prompt=%s",
			truncate(p, 800))
	}

	// Forbidden-only is insufficient: if Forbidden appears, polyfill word already required above.
	// Reject pure legacy wording that still says Target will fail with -32000 without polyfill path.
	if strings.Contains(p, "-32000") && strings.Contains(low, "not allowed") && !hasPolyfillWord {
		t.Fatalf("prompt still Forbidden/-32000 only without polyfill path; prompt=%s", truncate(p, 800))
	}
}
```
