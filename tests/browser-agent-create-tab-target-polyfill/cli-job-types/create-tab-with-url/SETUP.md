# Scenario

**Feature**: create-tab with positional URL posts params.url (B2)

```
Serve + fake WS
HandleCLI session create-tab --session-id X --addr <base> https://example.com/create-tab-marker
  -> observed type=create_tab
  -> params include url with example.com/create-tab-marker
  -> CLI nil error; stdout trailing \n
```

## Preconditions

- JobCLI = create-tab.
- CreateTabURL = https://example.com/create-tab-marker

## Steps

1. Set JobCLICreateTab.
2. Set CreateTabURL.
3. Leave CLIArgs empty for harness build.

## Context

- Requirement B2. Positional URL is enough (flag form optional in implementer).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobCLI = JobCLICreateTab
	req.CreateTabURL = "https://example.com/create-tab-marker"
	req.CLIArgs = nil
	return nil
}
```
