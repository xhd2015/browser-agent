## Expected

All paths exist relative to repo root:

- `har-viewer/main.go`
- `har-viewer/server/`
- `har-viewer/project-api-capture-react/`

## Side Effects

- None.

## Errors

- Any missing path fails the leaf.

## Exit Code

- Not asserted.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.RunErr != "" {
		t.Fatalf("layout probe: %s", resp.RunErr)
	}
	required := []string{
		"har-viewer/main.go",
		"har-viewer/server",
		"har-viewer/project-api-capture-react",
	}
	for _, rel := range required {
		if !resp.PathExists[rel] {
			t.Fatalf("missing required path %q under repo root %s", rel, req.RepoRoot)
		}
	}
}
```