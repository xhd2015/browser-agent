## Expected

- `browsertrace.Run` fails to listen on the occupied address.
- Process/result exit code is non-zero.
- Combined stderr / error text mentions that the address could not be bound
  (e.g. contains “in use”, “listen”, “address”, or “bind” — case-insensitive).

## Side Effects

- No successful session directory with final `recording.har` under `BaseDir`.
- Control server is not left serving on a fallback port.

## Errors

- Hard error only — no silent retry on another port.

## Exit Code

- Non-zero.

```go
import (
	"os"
	"path/filepath"
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
	ok := strings.Contains(low, "in use") ||
		strings.Contains(low, "listen") ||
		strings.Contains(low, "address already") ||
		strings.Contains(low, "bind") ||
		strings.Contains(low, "eaddrinuse")
	if !ok {
		t.Fatalf("error text should mention listen/address-in-use; got:\n%s", text)
	}
	if resp.HARPath != "" && harFileExists(resp.HARPath) {
		t.Fatalf("did not expect recording.har after bind failure: %s", resp.HARPath)
	}
	entries, _ := os.ReadDir(req.BaseDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		har := filepath.Join(req.BaseDir, e.Name(), "recording.har")
		if harFileExists(har) {
			t.Fatalf("unexpected recording.har after bind failure: %s", har)
		}
	}
}
```
