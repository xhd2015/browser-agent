# Scenario

**Bug**: human `session info` truncates enriched table after first tab and omits footer details

```
HandleCLI session info (default human) -> full Idx/ID/Active/Role/Title table + job_target + recommended_cli + session-page hint
```

## Preconditions

- InfoOp = human-table-columns.
- Fake extension returns two enriched tabs (111 session-page, 222 user active) plus job_target tab_index 2.

## Steps

1. Set `InfoOp = InfoOpHumanTableColumns`.

## Context

- Default human output (no --json). **RED** until `formatCompactEnrichedSessionInfo` loops all tabs and prints footer like `formatEnrichedTabsTable`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.InfoOp = InfoOpHumanTableColumns
	return nil
}
```