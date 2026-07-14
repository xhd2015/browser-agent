## Expected

- `WriteDaemonMeta` succeeds.
- `ReadDaemonMeta` succeeds and fields match written meta.
- Raw JSON file ends with `\n`.

## Side Effects

- Creates `server.json` under temp BaseDir.

## Errors

- Write or read error fails this leaf.

## Exit Code

- 0.

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
	if resp.WriteErr != nil {
		t.Fatalf("WriteDaemonMeta err=%v", resp.WriteErr)
	}
	if resp.ReadErr != nil {
		t.Fatalf("ReadDaemonMeta err=%v", resp.ReadErr)
	}
	daemonMetaFieldsEqual(t, resp.ReadMeta, req.Meta)
	if len(resp.ReadRawJSON) > 0 && !strings.HasSuffix(string(resp.ReadRawJSON), "\n") {
		t.Fatalf("server.json must end with trailing newline; tail=%q", string(resp.ReadRawJSON[len(resp.ReadRawJSON)-5:]))
	}
	assertExitZero(t, resp)
}
```