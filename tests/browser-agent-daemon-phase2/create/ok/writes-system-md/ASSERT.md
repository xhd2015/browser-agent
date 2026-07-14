## Expected

- `Create` succeeds.
- `SYSTEM.md` exists in session dir.
- File content equals `FormatSystemPrompt(sessionID)` byte-for-byte.
- `CreateSessionResult.SystemPath` points at the written file.

## Side Effects

- Creates `{baseDir}/sessions/playbook-test/SYSTEM.md`.

## Errors

- Missing file or content mismatch fails this leaf.

## Exit Code

- 0.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.CreateErr != nil {
		t.Fatalf("Create err=%v want nil", resp.CreateErr)
	}
	if !resp.SystemMDExists {
		t.Fatal("SYSTEM.md missing")
	}
	want := browseragent.FormatSystemPrompt(req.SessionID)
	if resp.SystemMDContent != want {
		t.Fatalf("SYSTEM.md content mismatch (len got %d want %d)", len(resp.SystemMDContent), len(want))
	}
	sysOnDisk := filepath.Join(browseragent.SessionDirPath(req.BaseDir, req.SessionID), "SYSTEM.md")
	if resp.CreateResult == nil || resp.CreateResult.SystemPath == "" {
		t.Fatal("CreateResult.SystemPath is empty")
	}
	if _, statErr := os.Stat(sysOnDisk); statErr != nil {
		t.Fatalf("SYSTEM.md on disk: %v", statErr)
	}
	assertExitZero(t, resp)
}
```