# Scenario

**Feature**: create-tab without URL posts job type create_tab (B1)

```
Serve + fake WS
HandleCLI session create-tab --session-id X --addr <base>
  -> observed type=create_tab
  -> CLI nil error; stdout trailing \n
```

## Preconditions

- JobCLI = create-tab.
- CreateTabURL empty (blank/new tab).

## Steps

1. Set JobCLICreateTab.
2. Clear CreateTabURL.
3. Leave CLIArgs empty for harness build.

## Context

- Requirement B1. Identity result shape asserted only via fake extension stub; CLI must accept.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobCLI = JobCLICreateTab
	req.CreateTabURL = ""
	req.CLIArgs = nil
	return nil
}
```
