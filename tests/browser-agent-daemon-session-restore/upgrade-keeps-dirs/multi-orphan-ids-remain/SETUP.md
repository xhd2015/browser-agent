# Scenario

**Feature**: Multiple orphan ids all keep their session dirs after kill+respawn

```
seed sess-o1 + sess-o2 + markers
  -> kill+respawn([sess-o1, sess-o2])
  -> both dirs + markers remain
```

## Preconditions

- UpgradeKeepCase = multi-orphan-ids-remain.
- OrphanIDs = [`sess-o1`, `sess-o2`].

## Steps

1. Set UpgradeKeepCase, OrphanIDs, Seeds.

## Context

- Wipe policy is all-or-nothing for the orphan id list; none may be deleted.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UpgradeKeepCase = UpgradeKeepMultiOrphan
	req.OrphanIDs = []string{"sess-o1", "sess-o2"}
	req.Seeds = []SeedSession{
		{ID: "sess-o1", Marker: "keep-me"},
		{ID: "sess-o2", Marker: "keep-me"},
	}
	return nil
}
```
