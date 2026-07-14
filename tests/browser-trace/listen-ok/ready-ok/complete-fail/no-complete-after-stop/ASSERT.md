## Expected

- Exit code ≠ 0 after complete-timeout.
- Error / meta indicates timeout or failure to receive final HAR.
- Final `recording.har` is either absent or not a truncated/corrupt JSON document
  (no overwrite with partial garbage). Prefer absent or previous valid-only policy.

## Side Effects

- Mock may have received `stop` command (`MockReceivedStop` ideally true once
  implementer queues stop on cancel).
- `meta.json` may exist documenting `stop_reason=timeout` / status `failed`.

## Errors

- Complete-phase timeout.

## Exit Code

- Non-zero.

```go
import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitNonZero(t, resp)
	text := combinedErrText(resp)
	low := strings.ToLower(text)
	hasSignal := strings.Contains(low, "timeout") ||
		strings.Contains(low, "complete") ||
		strings.Contains(low, "deadline") ||
		strings.Contains(low, "har")
	if !hasSignal {
		t.Fatalf("expected complete-timeout style message; got:\n%s", text)
	}

	// If meta exists, it should not claim a clean saved success without HAR.
	if len(resp.MetaJSON) > 0 {
		var meta map[string]any
		if err := json.Unmarshal(resp.MetaJSON, &meta); err != nil {
			t.Fatalf("meta.json is corrupt JSON: %v", err)
		}
		// If status field present, should not be "saved" without successful complete.
		if st, ok := meta["status"].(string); ok {
			if strings.EqualFold(st, "saved") {
				t.Fatalf("meta status=saved on complete timeout; meta=%s", resp.MetaJSON)
			}
		}
	}

	// recording.har: if present, must be valid JSON object (not truncated garbage).
	if resp.HARPath != "" {
		if st, err := os.Stat(resp.HARPath); err == nil && st.Size() > 0 {
			var v any
			if err := json.Unmarshal(resp.HARJSON, &v); err != nil {
				t.Fatalf("recording.har exists but is corrupt JSON (atomic write violated): %v\n%s", err, resp.HARJSON)
			}
			// Policy: complete timeout must not present a "final" success HAR as if complete won.
			// Implementers may omit the file entirely (preferred) or leave no success marker.
			t.Logf("note: recording.har present after complete timeout (%d bytes); ensure product does not treat session as saved", st.Size())
		}
	}

	if resp.ExitCode == 0 {
		t.Fatal("exit 0 on complete timeout")
	}
}
```
