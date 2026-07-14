## Expected

- Exit code 0.
- Mock observed a `start` command and a `stop` command (`MockReceivedStart`, `MockReceivedStop`).
- `recording.har` and `meta.json` written under session dir.
- Stdout ends with trailing `\n` when non-empty.

## Side Effects

- Session status ends as saved (meta may include `stop_reason=cli`).
- No complete-timeout failure.

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

	if !resp.MockReceivedStart {
		t.Fatal("mock did not receive start command")
	}
	if !resp.MockReceivedStop {
		t.Fatal("mock did not receive stop command after CLI cancel/signal")
	}
	if !resp.CompletePosted && len(resp.HARJSON) == 0 {
		// Prefer CompletePosted; HAR on disk is acceptable proof complete was processed.
		t.Fatal("expected complete to be posted and/or recording.har written")
	}

	if resp.SessionDir == "" {
		t.Fatal("SessionDir empty")
	}
	assertMetaPresent(t, resp.MetaJSON)
	assertHARHasMergedEntries(t, resp.HARJSON, 1)

	if len(resp.MetaJSON) > 0 {
		var meta map[string]any
		_ = json.Unmarshal(resp.MetaJSON, &meta)
		if sr, ok := meta["stop_reason"].(string); ok && sr != "" {
			if !strings.EqualFold(sr, "cli") && !strings.EqualFold(sr, "signal") {
				// Allow implementer synonym for CLI-initiated stop.
				t.Logf("meta.stop_reason = %q (expected cli/signal-ish)", sr)
			}
		}
	}

	if resp.Stdout != "" {
		assertStdoutTrailingNewline(t, resp.Stdout)
	}
}
```
