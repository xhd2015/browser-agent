# Scenario

**Feature**: `har inspect show --match` redacts Authorization

```
HandleCLI(show single-api.har --match /api/v1/items/create --json)
  -> authorization <REDACTED>; result.id 42
```

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	har := filepath.Join(req.ModuleRoot, "browseragent", "testdata", "har-inspect", "single-api.har")
	req.CLIArgs = []string{"har", "inspect", "show", har, "--match", "/api/v1/items/create", "--json"}
	req.CLIEnv = map[string]string{"NO_COLOR": "1"}
	return nil
}
```
