# Scenario

**Feature**: directory without manifest.json errors

```
HandleCLI(summary export-no-manifest) -> non-nil CLIErr mentioning manifest.json
```

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	dir := filepath.Join(req.ModuleRoot, "browseragent", "testdata", "har-inspect", "export-no-manifest")
	req.CLIArgs = []string{"har", "inspect", "summary", dir}
	req.CLIEnv = map[string]string{}
	return nil
}
```
