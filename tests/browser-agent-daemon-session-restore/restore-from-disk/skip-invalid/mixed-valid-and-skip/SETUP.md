# Scenario

**Feature**: Valid sessions restore while invalid siblings are skipped

```
seed sess-good (valid) + sess-badmeta (corrupt)
  -> RestoreSessionsFromDisk
  -> only sess-good registered; err nil
```

## Preconditions

- RestoreCase = mixed-valid-and-skip.
- Seeds: one valid, one corrupt.

## Steps

1. Set RestoreCase and mixed Seeds.

## Context

- Continue-on-error: one bad entry must not block good ones.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RestoreCase = RestoreCaseMixedValidAndSkip
	req.Seeds = []SeedSession{
		{ID: "sess-good"},
		{ID: "sess-badmeta", CorruptMeta: true},
	}
	return nil
}
```
