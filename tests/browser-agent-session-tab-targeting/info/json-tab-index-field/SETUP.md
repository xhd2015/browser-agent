# Scenario

**Feature**: --json includes tabs[].index and job_target.tab_index

```
HandleCLI session info --json -> tabs[].index + job_target.tab_index in stdout JSON
```

## Preconditions

- InfoOp = json-tab-index-field.

## Steps

1. Set `InfoOp = InfoOpJSONTabIndexField`.

## Context

- Fake extension returns tabs with index 1/2 and job_target.tab_index=2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.InfoOp = InfoOpJSONTabIndexField
	return nil
}
```