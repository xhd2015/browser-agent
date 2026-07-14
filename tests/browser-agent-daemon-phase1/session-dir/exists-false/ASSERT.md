## Expected

- `SessionDirExists` returns false when directory was not created.
- `SessionDirPath` still equals `{baseDir}/sessions/{sessionID}`.

## Side Effects

- None (no directory created).

## Errors

- Exists true fails this leaf.

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
	if resp.SessionDirExists {
		t.Fatalf("SessionDirExists(%q, %q)=true want false", req.BaseDir, req.SessionDirID)
	}
	wantPath := expectedSessionDirPath(req.BaseDir, req.SessionDirID)
	if resp.SessionDirPath != wantPath {
		t.Fatalf("SessionDirPath=%q want %q", resp.SessionDirPath, wantPath)
	}
	assertExitZero(t, resp)
}
```