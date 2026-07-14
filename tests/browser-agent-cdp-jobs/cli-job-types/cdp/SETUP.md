# Scenario

**Feature**: cdp posts job type cdp with Page.navigate method (B5)

```
Serve + fake WS
HandleCLI session cdp --session-id X --addr <base> Page.navigate {"url":"https://example.com"}
  -> observed type=cdp
  -> params.method = Page.navigate (or method field)
  -> CLI nil error; stdout trailing \n
```

## Preconditions

- JobCLI = cdp.
- CDPMethod = Page.navigate.
- CDPParamsJSON has example.com url.

## Steps

1. Set JobCLICdp.
2. Set method and params JSON.

## Context

- Requirement B5.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobCLI = JobCLICdp
	req.CDPMethod = "Page.navigate"
	req.CDPParamsJSON = `{"url":"https://example.com"}`
	req.CLIArgs = nil
	return nil
}
```
