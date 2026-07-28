# Scenario

**Feature**: Directory name that fails ValidateSessionID is skipped

```
seed sessions/-badid/meta.json -> RestoreSessionsFromDisk
  -> not registered; err nil
```

## Preconditions

- RestoreCase = bad-session-id-dirname.
- Seed id `-badid` (leading dash — invalid session id).

## Steps

1. Set RestoreCase and Seeds with invalid dirname.

## Context

- Dirname is the session id candidate; validation failure → skip.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RestoreCase = RestoreCaseBadSessionIDDir
	req.Seeds = []SeedSession{
		{ID: "-badid"},
	}
	return nil
}
```
