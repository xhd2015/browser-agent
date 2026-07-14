# Scenario

**Feature**: screenshot posts job type screenshot (B4)

```
Serve + fake WS
HandleCLI session screenshot --session-id X --addr <base>
  -> observed type=screenshot
  -> CLI nil error; stdout trailing \n
```

## Preconditions

- JobCLI = screenshot.
- No `-o` required for this leaf.

## Steps

1. Set JobCLIScreenshot.

## Context

- Requirement B4. Optional format/full_page params not required.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobCLI = JobCLIScreenshot
	req.CLIArgs = nil
	return nil
}
```
