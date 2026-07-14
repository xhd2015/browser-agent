## Expected Output

Stdout prints the session/out path and **ends with a trailing newline** (POSIX).
Exact banner lines may vary; path must appear and final byte must be `\n`.

```
---
version: 2
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-abc, session directory path
---
__PATH__
```

(Implementers may print a short label line before the path; Assert accepts either a
single path line or multi-line stdout as long as the path is present and `\n` ends stdout.)

## Expected

- Exit code 0.
- `recording.har` and `meta.json` exist under the session directory.
- HAR contains ≥2 merged entries (multi-tab); URLs from both tabs present when using fixture.
- Session directory name matches `YYYY-MM-DD-HH-MM-SS-<suffix>` under `BaseDir`.
- Stdout includes the session path and ends with `\n`.

## Side Effects

- Files written under `{BaseDir}/{session-name}/` only (not scattered).
- `stop_reason` in meta is `extension` when present.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitZero(t, resp)

	if resp.SessionDir == "" {
		t.Fatal("SessionDir empty after successful run")
	}
	assertSessionDirPattern(t, resp.SessionDir, req.BaseDir)

	assertMetaPresent(t, resp.MetaJSON)
	assertHARHasMergedEntries(t, resp.HARJSON, 2)

	// Multi-tab: fixture/default includes example.com/a and /b
	harStr := string(resp.HARJSON)
	if !strings.Contains(harStr, "example.com/a") || !strings.Contains(harStr, "example.com/b") {
		t.Fatalf("HAR should include multi-tab URLs example.com/a and /b; got:\n%s", harStr)
	}

	// Prefer entries sorted by startedDateTime (a before b).
	var doc struct {
		Log struct {
			Entries []struct {
				StartedDateTime string `json:"startedDateTime"`
				Request         struct {
					URL string `json:"url"`
				} `json:"request"`
			} `json:"entries"`
		} `json:"log"`
	}
	if err := json.Unmarshal(resp.HARJSON, &doc); err != nil {
		t.Fatalf("HAR parse: %v", err)
	}
	if len(doc.Log.Entries) >= 2 {
		// If product sorts, first startedDateTime <= second.
		if doc.Log.Entries[0].StartedDateTime > doc.Log.Entries[1].StartedDateTime {
			t.Fatalf("HAR entries not sorted by startedDateTime: %q then %q",
				doc.Log.Entries[0].StartedDateTime, doc.Log.Entries[1].StartedDateTime)
		}
	}

	if len(resp.MetaJSON) > 0 {
		var meta map[string]any
		_ = json.Unmarshal(resp.MetaJSON, &meta)
		if sr, ok := meta["stop_reason"].(string); ok && sr != "" && !strings.EqualFold(sr, "extension") {
			t.Fatalf("meta.stop_reason = %q, want extension", sr)
		}
	}

	// Stdout: path present + trailing newline (#8).
	assertStdoutTrailingNewline(t, resp.Stdout)
	if !strings.Contains(resp.Stdout, resp.SessionDir) && !strings.Contains(resp.Stdout, filepathBase(resp.SessionDir)) {
		// Accept either absolute session dir or at least the leaf name.
		t.Fatalf("stdout should mention session path %q; stdout=%q", resp.SessionDir, resp.Stdout)
	}
}

func filepathBase(p string) string {
	p = strings.TrimRight(p, "/\\")
	i := strings.LastIndexAny(p, `/\\`)
	if i < 0 {
		return p
	}
	return p[i+1:]
}
```
