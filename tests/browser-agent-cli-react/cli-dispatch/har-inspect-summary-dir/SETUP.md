# Scenario

**Feature**: `har inspect summary` on export dir with manifest

```
HandleCLI(["har","inspect","summary", <export-ok>, "--host","app.example.com"])
  -> session id + tab files; exit 0
```

## Steps

1. Resolve fixture under ModuleRoot/browseragent/testdata/har-inspect/export-ok
2. Invoke summary with host filter

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	dir := filepath.Join(req.ModuleRoot, "browseragent", "testdata", "har-inspect", "export-ok")
	req.CLIArgs = []string{"har", "inspect", "summary", dir, "--host", "app.example.com"}
	req.CLIEnv = map[string]string{"NO_COLOR": "1"}
	return nil
}
```
