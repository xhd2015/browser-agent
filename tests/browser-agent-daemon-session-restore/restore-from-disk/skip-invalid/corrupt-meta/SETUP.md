# Scenario

**Feature**: Corrupt meta.json is skipped without failing restore

```
seed sessions/sess-badmeta/meta.json = "{not-json" -> RestoreSessionsFromDisk
  -> not registered; err nil
```

## Preconditions

- RestoreCase = corrupt-meta.
- Seed `sess-badmeta` with CorruptMeta true.

## Steps

1. Set RestoreCase and Seeds with CorruptMeta.

## Context

- Unparseable JSON is skip, not hard error.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RestoreCase = RestoreCaseCorruptMeta
	req.Seeds = []SeedSession{
		{ID: "sess-badmeta", CorruptMeta: true},
	}
	return nil
}
```
