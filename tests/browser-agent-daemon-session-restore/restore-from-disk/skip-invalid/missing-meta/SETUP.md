# Scenario

**Feature**: Session dir without meta.json is skipped

```
seed sessions/sess-nomete/ (no meta.json) -> RestoreSessionsFromDisk
  -> not registered; err nil
```

## Preconditions

- RestoreCase = missing-meta.
- Seed `sess-nomete` with NoMeta true.

## Steps

1. Set RestoreCase and Seeds with NoMeta.

## Context

- Directory alone is insufficient; meta.json required.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RestoreCase = RestoreCaseMissingMeta
	req.Seeds = []SeedSession{
		{ID: "sess-nomete", NoMeta: true},
	}
	return nil
}
```
