# Scenario

**Feature**: Disk-seeded session appears in GET /v1/sessions after RunDaemon

```
seed sess-wire01/meta.json -> RunDaemon -> GET /v1/sessions includes sess-wire01
```

## Preconditions

- RunDaemonWireCase = list-includes-disk.
- Seed `sess-wire01` with valid meta.

## Steps

1. Set RunDaemonWireCase and Seeds.

## Context

- End-to-end proof that restore is wired into RunDaemon (not only the pure helper).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.RunDaemonWireCase = RunDaemonListIncludesDisk
	req.Seeds = []SeedSession{
		{ID: "sess-wire01"},
	}
	return nil
}
```
