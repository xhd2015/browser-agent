# Scenario

**Feature**: successful Create writes session artifacts

```
Create(id) -> mkdir + meta.json + SYSTEM.md + registry entry
```

## Preconditions

- CreateCase is ok.
- SessionID is a valid id.

## Steps

1. Set CreateCase to ok.

## Context

- Artifacts live under `{baseDir}/sessions/{id}/`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CreateCase = CreateCaseOK
	if req.SessionID == "" {
		req.SessionID = "my-flow"
	}
	return nil
}
```