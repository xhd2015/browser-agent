## Expected

- `Create("my-flow")` succeeds with non-nil `CreateSessionResult`.
- `meta.json` exists under session dir.
- Parsed fields: `session_id`, `addr`, `base_url`, `session_url`, `system_prompt_path`, `product`, `control_port` match registry addr and id.
- `CreateSessionResult.SessionURL` equals `http://127.0.0.1:43761/go?session=my-flow`.

## Side Effects

- Creates `{baseDir}/sessions/my-flow/meta.json`.

## Errors

- Create error or missing/mismatched meta fields fail this leaf.

## Exit Code

- 0.

```go
import (
	"os"
	"path/filepath"
	"strings"
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
	if resp.CreateResult == nil {
		t.Fatal("CreateResult is nil")
	}
	if !resp.MetaFileExists {
		t.Fatal("meta.json missing")
	}
	if resp.MetaSessionID != req.SessionID {
		t.Fatalf("meta session_id=%q want %q", resp.MetaSessionID, req.SessionID)
	}
	if resp.MetaAddr != req.Addr {
		t.Fatalf("meta addr=%q want %q", resp.MetaAddr, req.Addr)
	}
	wantBase := expectedBaseURL(req.Addr)
	if resp.MetaBaseURL != wantBase {
		t.Fatalf("meta base_url=%q want %q", resp.MetaBaseURL, wantBase)
	}
	wantURL := expectedSessionURL(req.Addr, req.SessionID)
	if resp.MetaSessionURL != wantURL {
		t.Fatalf("meta session_url=%q want %q", resp.MetaSessionURL, wantURL)
	}
	if resp.CreateResult.SessionURL != wantURL {
		t.Fatalf("CreateResult.SessionURL=%q want %q", resp.CreateResult.SessionURL, wantURL)
	}
	if resp.MetaProduct != "browser-agent" {
		t.Fatalf("meta product=%q want browser-agent", resp.MetaProduct)
	}
	wantPort := expectedControlPort(req.Addr)
	if resp.MetaControlPort != wantPort {
		t.Fatalf("meta control_port=%d want %d", resp.MetaControlPort, wantPort)
	}
	if resp.MetaSystemPromptPath == "" {
		t.Fatal("meta system_prompt_path is empty")
	}
	if !strings.HasSuffix(resp.MetaSystemPromptPath, "SYSTEM.md") {
		t.Fatalf("system_prompt_path=%q should end with SYSTEM.md", resp.MetaSystemPromptPath)
	}
	wantDir := browseragent.SessionDirPath(req.BaseDir, req.SessionID)
	if !browseragent.SessionDirExists(req.BaseDir, req.SessionID) {
		t.Fatalf("session dir missing under %q", wantDir)
	}
	if resp.CreateResult.SessionDir == "" {
		t.Fatal("CreateResult.SessionDir is empty")
	}
	if !filepath.IsAbs(resp.CreateResult.SessionDir) {
		t.Fatalf("CreateResult.SessionDir=%q want absolute path", resp.CreateResult.SessionDir)
	}
	metaOnDisk := filepath.Join(wantDir, "meta.json")
	if resp.CreateResult.MetaPath == "" {
		t.Fatal("CreateResult.MetaPath is empty")
	}
	if _, statErr := os.Stat(metaOnDisk); statErr != nil {
		t.Fatalf("meta on disk: %v", statErr)
	}
	assertExitZero(t, resp)
}
```