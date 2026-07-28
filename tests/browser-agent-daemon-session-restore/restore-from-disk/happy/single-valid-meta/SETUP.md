# Scenario

**Feature**: Single valid session dir with meta.json restores into registry

```
seed sessions/sess-alpha/meta.json -> RestoreSessionsFromDisk
  -> Get(sess-alpha)=true; phase waiting_extension
```

## Preconditions

- RestoreCase = single-valid-meta.
- One seed: `sess-alpha` with valid meta.json.

## Steps

1. Set RestoreCase and Seeds.

## Context

- Canonical happy path for Phase 1 restore.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RestoreCase = RestoreCaseSingleValidMeta
	req.Seeds = []SeedSession{
		{ID: "sess-alpha"},
	}
	return nil
}
```
