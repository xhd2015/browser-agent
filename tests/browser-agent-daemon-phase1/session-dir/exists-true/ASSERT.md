## Expected

- `SessionDirExists` returns true after `MkdirAll`.
- `SessionDirPath` equals `{baseDir}/sessions/{sessionID}`.

## Side Effects

- Creates session directory under temp BaseDir.

## Errors

- Exists false or path mismatch fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.SessionDirExists {
		t.Fatalf("SessionDirExists(%q, %q)=false want true", req.BaseDir, req.SessionDirID)
	}
	wantPath := expectedSessionDirPath(req.BaseDir, req.SessionDirID)
	if resp.SessionDirPath != wantPath {
		t.Fatalf("SessionDirPath=%q want %q", resp.SessionDirPath, wantPath)
	}
	assertExitZero(t, resp)
}
```