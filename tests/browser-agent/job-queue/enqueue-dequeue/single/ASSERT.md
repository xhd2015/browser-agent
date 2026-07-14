## Expected

Requirement **B1**:

- Exactly one dequeued id (non-empty).
- Dequeued type is `eval`.
- Params round-trip includes `code` = `1+1` (via `JobResultData` snapshot from Run).

## Side Effects

- None beyond in-memory queue.

## Errors

- Empty id or wrong type fails the leaf.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if len(resp.DequeuedIDs) != 1 || resp.DequeuedIDs[0] == "" {
		t.Fatalf("DequeuedIDs = %v, want one non-empty id", resp.DequeuedIDs)
	}
	if len(resp.DequeuedTypes) != 1 || resp.DequeuedTypes[0] != "eval" {
		t.Fatalf("DequeuedTypes = %v, want [eval]", resp.DequeuedTypes)
	}
	if resp.JobResultData == nil {
		t.Fatal("JobResultData snapshot missing")
	}
	if typ, _ := resp.JobResultData["type"].(string); typ != "eval" {
		t.Fatalf("snapshot type = %v, want eval", resp.JobResultData["type"])
	}
	params, _ := resp.JobResultData["params"].(map[string]any)
	if params == nil {
		t.Fatalf("params not preserved; snapshot=%v", resp.JobResultData)
	}
	if code, _ := params["code"].(string); code != "1+1" {
		t.Fatalf("params.code = %v, want 1+1", params["code"])
	}
}
```
